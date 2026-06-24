package gormdb

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

var _ repository.ComponentRepository = (*componentRepository)(nil)

type componentRepository struct {
	database *gorm.DB
}

func NewComponentRepository(database *gorm.DB) repository.ComponentRepository {
	return &componentRepository{database: database}
}

func (r *componentRepository) table(slug, component string) *gorm.DB {
	return r.database.Table(componentTableName(slug, component))
}

func (r *componentRepository) isPostgres() bool {
	return r.database.Dialector.Name() == "postgres"
}

func (r *componentRepository) EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition, isNested bool) error {
	table := componentTableName(contentTypeSlug, componentName)
	if !r.database.Migrator().HasTable(table) {
		return r.createComponentTable(ctx, table, fields, isNested)
	}
	return r.addMissingComponentColumns(ctx, table, fields, isNested)
}

func (r *componentRepository) createComponentTable(ctx context.Context, table string, fields []entity.FieldDefinition, isNested bool) error {
	var cols []string
	if r.isPostgres() {
		cols = append(cols, "gorm_id SERIAL PRIMARY KEY")
	} else {
		cols = append(cols, "gorm_id INTEGER PRIMARY KEY AUTOINCREMENT")
	}
	cols = append(cols, "component_id TEXT")
	if isNested {
		cols = append(cols, "parent_component_id TEXT")
	} else {
		cols = append(cols, "document_id TEXT")
	}
	cols = append(cols, "version TEXT")
	cols = append(cols, "locale TEXT")
	cols = append(cols, "sort_order INTEGER DEFAULT 0")
	for _, field := range flattenLayoutFields(fields) {
		if field.Type == "component" {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s %s", toSnakeCase(field.Name), fieldColumnType(field.Type)))
	}
	cols = append(cols, "created_at TIMESTAMP")
	cols = append(cols, "updated_at TIMESTAMP")

	sql := fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(cols, ", "))
	return r.database.WithContext(ctx).Exec(sql).Error
}

func (r *componentRepository) addMissingComponentColumns(ctx context.Context, table string, fields []entity.FieldDefinition, isNested bool) error {
	cols, err := existingColumns(r.database, table)
	if err != nil {
		return err
	}
	if !cols["sort_order"] {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN sort_order INTEGER DEFAULT 0", table)
		if err := r.database.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	if isNested && !cols["parent_component_id"] {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN parent_component_id TEXT", table)
		if err := r.database.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	if !isNested && !cols["document_id"] {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN document_id TEXT", table)
		if err := r.database.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	for _, field := range flattenLayoutFields(fields) {
		if field.Type == "component" {
			continue
		}
		col := toSnakeCase(field.Name)
		if cols[col] {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, fieldColumnType(field.Type))
		if err := r.database.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *componentRepository) DropCollection(ctx context.Context, contentTypeSlug, componentName string) error {
	table := componentTableName(contentTypeSlug, componentName)
	return r.database.WithContext(ctx).Migrator().DropTable(table)
}

func compToRow(comp *entity.Component) map[string]any {
	row := map[string]any{
		"component_id": comp.ComponentID,
		"version":      string(comp.Version),
		"locale":       comp.Locale,
		"sort_order":   comp.SortOrder,
		"created_at":   comp.CreatedAt,
		"updated_at":   comp.UpdatedAt,
	}
	if comp.ParentComponentID != "" {
		row["parent_component_id"] = comp.ParentComponentID
	} else {
		row["document_id"] = comp.DocumentID
	}
	for key, val := range comp.Fields {
		row[toSnakeCase(key)] = serializeFieldValue(val)
	}
	return row
}

func rowToComp(row map[string]any) *entity.Component {
	comp := &entity.Component{}
	if val, ok := row["gorm_id"]; ok {
		comp.GormID = toUint(val)
	}
	if val, ok := row["component_id"]; ok {
		comp.ComponentID = toString(val)
	}
	if val, ok := row["document_id"]; ok {
		comp.DocumentID = toString(val)
	}
	if val, ok := row["parent_component_id"]; ok {
		comp.ParentComponentID = toString(val)
	}
	if val, ok := row["version"]; ok {
		comp.Version = entity.DocumentVersion(toString(val))
	}
	if val, ok := row["locale"]; ok {
		comp.Locale = toString(val)
	}
	if val, ok := row["sort_order"]; ok {
		comp.SortOrder = toInt(val)
	}
	if val, ok := row["created_at"]; ok {
		comp.CreatedAt = toTime(val)
	}
	if val, ok := row["updated_at"]; ok {
		comp.UpdatedAt = toTime(val)
	}

	systemCols := map[string]bool{
		"gorm_id": true, "component_id": true, "document_id": true, "parent_component_id": true,
		"version": true, "locale": true, "sort_order": true, "created_at": true, "updated_at": true,
	}
	fields := make(map[string]any)
	for key, val := range row {
		if !systemCols[key] {
			fields[toCamelCase(key)] = deserializeFieldValue(val)
		}
	}
	if len(fields) > 0 {
		comp.Fields = fields
	}
	return comp
}

func (r *componentRepository) FindByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error) {
	var rows []map[string]any
	err := r.table(contentTypeSlug, componentName).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
		Order("sort_order ASC, gorm_id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	components := make([]*entity.Component, len(rows))
	for idx, row := range rows {
		components[idx] = rowToComp(row)
	}
	return components, nil
}

func (r *componentRepository) FindByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion) ([]*entity.Component, error) {
	var rows []map[string]any
	err := r.table(contentTypeSlug, componentPath).WithContext(ctx).
		Where("parent_component_id = ? AND version = ? AND locale = ?", parentComponentID, version, locale).
		Order("sort_order ASC, gorm_id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	components := make([]*entity.Component, len(rows))
	for idx, row := range rows {
		components[idx] = rowToComp(row)
	}
	return components, nil
}

func (r *componentRepository) UpsertAll(ctx context.Context, contentTypeSlug, componentName, documentID, locale string, version entity.DocumentVersion, components []*entity.Component) error {
	tbl := r.table(contentTypeSlug, componentName).WithContext(ctx)

	if err := tbl.
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
		Delete(map[string]any{}).Error; err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	for _, comp := range components {
		comp.DocumentID = documentID
		comp.ParentComponentID = ""
		comp.Version = version
		comp.Locale = locale
		row := compToRow(comp)
		if err := r.table(contentTypeSlug, componentName).WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *componentRepository) UpsertAllByParent(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string, version entity.DocumentVersion, components []*entity.Component) error {
	tbl := r.table(contentTypeSlug, componentPath).WithContext(ctx)

	if err := tbl.
		Where("parent_component_id = ? AND version = ? AND locale = ?", parentComponentID, version, locale).
		Delete(map[string]any{}).Error; err != nil {
		return err
	}

	if len(components) == 0 {
		return nil
	}

	for _, comp := range components {
		comp.ParentComponentID = parentComponentID
		comp.DocumentID = ""
		comp.Version = version
		comp.Locale = locale
		row := compToRow(comp)
		if err := r.table(contentTypeSlug, componentPath).WithContext(ctx).Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *componentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, componentName, documentID, locale string) error {
	return r.table(contentTypeSlug, componentName).WithContext(ctx).
		Where("document_id = ? AND locale = ?", documentID, locale).
		Delete(map[string]any{}).Error
}

func (r *componentRepository) DeleteByParentComponentID(ctx context.Context, contentTypeSlug, componentPath, parentComponentID, locale string) error {
	return r.table(contentTypeSlug, componentPath).WithContext(ctx).
		Where("parent_component_id = ? AND locale = ?", parentComponentID, locale).
		Delete(map[string]any{}).Error
}

func (r *componentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error {
	return r.table(contentTypeSlug, componentName).WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(map[string]any{}).Error
}
