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

func (r *componentRepository) EnsureCollection(ctx context.Context, contentTypeSlug, componentName string, fields []entity.FieldDefinition) error {
	table := componentTableName(contentTypeSlug, componentName)
	if !r.database.Migrator().HasTable(table) {
		return r.createComponentTable(ctx, table, fields)
	}
	return r.addMissingComponentColumns(ctx, table, fields)
}

func (r *componentRepository) createComponentTable(ctx context.Context, table string, fields []entity.FieldDefinition) error {
	var cols []string
	if r.isPostgres() {
		cols = append(cols, "gorm_id SERIAL PRIMARY KEY")
	} else {
		cols = append(cols, "gorm_id INTEGER PRIMARY KEY AUTOINCREMENT")
	}
	cols = append(cols, "component_id TEXT")
	cols = append(cols, "document_id TEXT")
	cols = append(cols, "version TEXT")
	cols = append(cols, "locale TEXT")
	cols = append(cols, "sort_order INTEGER DEFAULT 0")
	for _, f := range flattenLayoutFields(fields) {
		if f.Type == "component" {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s %s", toSnakeCase(f.Name), fieldColumnType(f.Type)))
	}
	cols = append(cols, "created_at TIMESTAMP")
	cols = append(cols, "updated_at TIMESTAMP")

	sql := fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(cols, ", "))
	return r.database.WithContext(ctx).Exec(sql).Error
}

func (r *componentRepository) addMissingComponentColumns(ctx context.Context, table string, fields []entity.FieldDefinition) error {
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
	for _, f := range flattenLayoutFields(fields) {
		if f.Type == "component" {
			continue
		}
		col := toSnakeCase(f.Name)
		if cols[col] {
			continue
		}
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, fieldColumnType(f.Type))
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

func compToRow(c *entity.Component) map[string]any {
	row := map[string]any{
		"component_id": c.ComponentID,
		"document_id":  c.DocumentID,
		"version":      string(c.Version),
		"locale":       c.Locale,
		"sort_order":   c.SortOrder,
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	}
	for k, v := range c.Fields {
		row[toSnakeCase(k)] = serializeFieldValue(v)
	}
	return row
}

func rowToComp(row map[string]any) *entity.Component {
	comp := &entity.Component{}
	if v, ok := row["gorm_id"]; ok {
		comp.GormID = toUint(v)
	}
	if v, ok := row["component_id"]; ok {
		comp.ComponentID = toString(v)
	}
	if v, ok := row["document_id"]; ok {
		comp.DocumentID = toString(v)
	}
	if v, ok := row["version"]; ok {
		comp.Version = entity.DocumentVersion(toString(v))
	}
	if v, ok := row["locale"]; ok {
		comp.Locale = toString(v)
	}
	if v, ok := row["sort_order"]; ok {
		comp.SortOrder = toInt(v)
	}
	if v, ok := row["created_at"]; ok {
		comp.CreatedAt = toTime(v)
	}
	if v, ok := row["updated_at"]; ok {
		comp.UpdatedAt = toTime(v)
	}

	systemCols := map[string]bool{
		"gorm_id": true, "component_id": true, "document_id": true,
		"version": true, "locale": true, "sort_order": true, "created_at": true, "updated_at": true,
	}
	fields := make(map[string]any)
	for k, v := range row {
		if !systemCols[k] {
			fields[toCamelCase(k)] = deserializeFieldValue(v)
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
	for i, row := range rows {
		components[i] = rowToComp(row)
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

	for _, c := range components {
		c.DocumentID = documentID
		c.Version = version
		c.Locale = locale
		row := compToRow(c)
		if err := r.table(contentTypeSlug, componentName).WithContext(ctx).Create(row).Error; err != nil {
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

func (r *componentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug, componentName string) error {
	return r.table(contentTypeSlug, componentName).WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(map[string]any{}).Error
}
