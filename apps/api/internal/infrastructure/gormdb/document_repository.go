package gormdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
	"project-abyssoftime-cms-v2/api/internal/domain/repository"
	pkgerrors "project-abyssoftime-cms-v2/api/pkg/errors"
)

var gormSortColumn = map[string]string{
	"id":        "gorm_id",
	"createdAt": "created_at",
	"updatedAt": "updated_at",
}

func resolveGormSortClause(orderBy string, sortDir int) string {
	col, ok := gormSortColumn[orderBy]
	if !ok {
		col = "created_at"
	}
	dir := "DESC"
	if sortDir == 1 {
		dir = "ASC"
	}
	return fmt.Sprintf("%s %s", col, dir)
}

var _ repository.DocumentRepository = (*documentRepository)(nil)

type documentRepository struct {
	database *gorm.DB
}

func NewDocumentRepository(database *gorm.DB) repository.DocumentRepository {
	return &documentRepository{database: database}
}

func (r *documentRepository) table(slug string) *gorm.DB {
	return r.database.Table(documentTableName(slug))
}

func existingColumns(db *gorm.DB, table string) (map[string]bool, error) {
	rows, err := db.Table(table).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(cols))
	for _, c := range cols {
		set[c] = true
	}
	return set, nil
}

func flattenLayoutFields(fields []entity.FieldDefinition) []entity.FieldDefinition {
	var result []entity.FieldDefinition
	for _, field := range fields {
		if field.Type == "layout" {
			result = append(result, field.Fields...)
		} else {
			result = append(result, field)
		}
	}
	return result
}

func fieldColumnType(fieldType string) string {
	switch fieldType {
	case "number":
		return "REAL"
	case "boolean":
		return "BOOLEAN"
	case "json":
		return "TEXT"
	default:
		return "TEXT"
	}
}

func (r *documentRepository) isPostgres() bool {
	return r.database.Dialector.Name() == "postgres"
}

func (r *documentRepository) EnsureCollection(ctx context.Context, contentTypeSlug string, fields []entity.FieldDefinition) error {
	table := documentTableName(contentTypeSlug)
	if !r.database.Migrator().HasTable(table) {
		return r.createDocumentTable(ctx, table, fields)
	}
	return r.addMissingDocumentColumns(ctx, table, fields)
}

func (r *documentRepository) createDocumentTable(ctx context.Context, table string, fields []entity.FieldDefinition) error {
	var cols []string
	if r.isPostgres() {
		cols = append(cols, "gorm_id SERIAL PRIMARY KEY")
	} else {
		cols = append(cols, "gorm_id INTEGER PRIMARY KEY AUTOINCREMENT")
	}
	cols = append(cols, "document_id TEXT")
	cols = append(cols, "version TEXT")
	cols = append(cols, "locale TEXT")
	for _, f := range flattenLayoutFields(fields) {
		if f.Type == "component" {
			continue
		}
		cols = append(cols, fmt.Sprintf("%s %s", toSnakeCase(f.Name), fieldColumnType(f.Type)))
	}
	cols = append(cols, "created_at TIMESTAMP")
	cols = append(cols, "updated_at TIMESTAMP")
	cols = append(cols, "published_at TIMESTAMP")
	cols = append(cols, "created_by TEXT")
	cols = append(cols, "updated_by TEXT")
	cols = append(cols, "published_by TEXT")

	sql := fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(cols, ", "))
	return r.database.WithContext(ctx).Exec(sql).Error
}

func (r *documentRepository) addMissingDocumentColumns(ctx context.Context, table string, fields []entity.FieldDefinition) error {
	cols, err := existingColumns(r.database, table)
	if err != nil {
		return err
	}
	flat := flattenLayoutFields(fields)
	for _, f := range flat {
		if f.Type == "component" {
			continue
		}
		col := toSnakeCase(f.Name)
		if cols[col] {
			continue
		}
		stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, col, fieldColumnType(f.Type))
		if err := r.database.WithContext(ctx).Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *documentRepository) DropCollection(ctx context.Context, contentTypeSlug string) error {
	table := documentTableName(contentTypeSlug)
	return r.database.WithContext(ctx).Migrator().DropTable(table)
}

func (r *documentRepository) TableInfo(ctx context.Context, contentTypeSlug string) (bool, int64, error) {
	table := documentTableName(contentTypeSlug)
	if !r.database.Migrator().HasTable(table) {
		return false, 0, nil
	}
	var count int64
	if err := r.database.WithContext(ctx).Table(table).Count(&count).Error; err != nil {
		return true, 0, err
	}
	return true, count, nil
}

func docToRow(doc *entity.Document) map[string]any {
	row := map[string]any{
		"document_id":  doc.DocumentID,
		"version":      string(doc.Version),
		"locale":       doc.Locale,
		"created_at":   doc.CreatedAt,
		"updated_at":   doc.UpdatedAt,
		"published_at": nil,
		"created_by":   doc.CreatedBy,
		"updated_by":   doc.UpdatedBy,
		"published_by": doc.PublishedBy,
	}
	if doc.PublishedAt != nil {
		row["published_at"] = *doc.PublishedAt
	}
	for k, v := range doc.Fields {
		row[toSnakeCase(k)] = serializeFieldValue(v)
	}
	return row
}

func serializeFieldValue(v any) any {
	if v == nil {
		return nil
	}
	switch v.(type) {
	case map[string]any, []any:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return v
	}
}

func deserializeFieldValue(v any) any {
	s, ok := v.(string)
	if !ok || len(s) == 0 {
		return v
	}
	if (s[0] == '{' && s[len(s)-1] == '}') || (s[0] == '[' && s[len(s)-1] == ']') {
		var parsed any
		if json.Unmarshal([]byte(s), &parsed) == nil {
			return parsed
		}
	}
	return v
}

func rowToDoc(row map[string]any) *entity.Document {
	doc := &entity.Document{}
	if v, ok := row["gorm_id"]; ok {
		doc.GormID = toUint(v)
	}
	if v, ok := row["document_id"]; ok {
		doc.DocumentID = toString(v)
	}
	if v, ok := row["version"]; ok {
		doc.Version = entity.DocumentVersion(toString(v))
	}
	if v, ok := row["locale"]; ok {
		doc.Locale = toString(v)
	}
	if v, ok := row["created_at"]; ok {
		doc.CreatedAt = toTime(v)
	}
	if v, ok := row["updated_at"]; ok {
		doc.UpdatedAt = toTime(v)
	}
	if v, ok := row["published_at"]; ok && v != nil {
		t := toTime(v)
		if !t.IsZero() {
			doc.PublishedAt = &t
		}
	}
	if v, ok := row["created_by"]; ok {
		doc.CreatedBy = toString(v)
	}
	if v, ok := row["updated_by"]; ok {
		doc.UpdatedBy = toString(v)
	}
	if v, ok := row["published_by"]; ok {
		doc.PublishedBy = toString(v)
	}

	systemCols := map[string]bool{
		"gorm_id": true, "document_id": true, "version": true, "locale": true,
		"created_at": true, "updated_at": true, "published_at": true,
		"created_by": true, "updated_by": true, "published_by": true,
	}
	fields := make(map[string]any)
	for k, v := range row {
		if !systemCols[k] {
			fields[toCamelCase(k)] = deserializeFieldValue(v)
		}
	}
	if len(fields) > 0 {
		doc.Fields = fields
	}
	return doc
}

func (r *documentRepository) findOne(ctx context.Context, slug, documentID, locale string, version entity.DocumentVersion) (*entity.Document, error) {
	var result map[string]any
	err := r.table(slug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?", documentID, version, locale).
		Take(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, err
	}
	return rowToDoc(result), nil
}

func (r *documentRepository) FindDraftByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findOne(ctx, contentTypeSlug, documentID, locale, entity.VersionDraft)
}

func (r *documentRepository) FindPublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) (*entity.Document, error) {
	return r.findOne(ctx, contentTypeSlug, documentID, locale, entity.VersionPublished)
}

func (r *documentRepository) upsert(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	var existing map[string]any
	err := r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?",
			doc.DocumentID, doc.Version, doc.Locale).
		Take(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row := docToRow(doc)
			return r.table(contentTypeSlug).WithContext(ctx).Create(row).Error
		}
		return err
	}
	doc.GormID = toUint(existing["gorm_id"])
	row := docToRow(doc)
	row["gorm_id"] = doc.GormID
	return r.table(contentTypeSlug).WithContext(ctx).Where("gorm_id = ?", doc.GormID).Updates(row).Error
}

func (r *documentRepository) UpsertDraft(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	doc.Version = entity.VersionDraft
	return r.upsert(ctx, contentTypeSlug, doc)
}

func (r *documentRepository) UpsertPublished(ctx context.Context, contentTypeSlug string, doc *entity.Document) error {
	doc.Version = entity.VersionPublished
	return r.upsert(ctx, contentTypeSlug, doc)
}

func (r *documentRepository) findMany(ctx context.Context, slug string, query *gorm.DB) ([]*entity.Document, error) {
	var rows []map[string]any
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	docs := make([]*entity.Document, len(rows))
	for i, row := range rows {
		docs[i] = rowToDoc(row)
	}
	return docs, nil
}

func (r *documentRepository) FindDraftsByContentType(ctx context.Context, contentTypeSlug string) ([]*entity.Document, error) {
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ?", entity.VersionDraft).
		Order("created_at DESC")
	return r.findMany(ctx, contentTypeSlug, q)
}

func (r *documentRepository) FindDraftsByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	var total int64
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ?", entity.VersionDraft, locale)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []map[string]any
	if err := q.Order(resolveGormSortClause(orderBy, sortDir)).Offset(start).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	docs := make([]*entity.Document, len(rows))
	for i, row := range rows {
		docs[i] = rowToDoc(row)
	}
	return docs, total, nil
}

func (r *documentRepository) FindPublishedByContentTypePaginated(ctx context.Context, contentTypeSlug string, start, size int, locale, orderBy string, sortDir int) ([]*entity.Document, int64, error) {
	var total int64
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ?", entity.VersionPublished, locale)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []map[string]any
	if err := q.Order(resolveGormSortClause(orderBy, sortDir)).Offset(start).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	docs := make([]*entity.Document, len(rows))
	for i, row := range rows {
		docs[i] = rowToDoc(row)
	}
	return docs, total, nil
}

func (r *documentRepository) FindPublishedByDocumentIDs(ctx context.Context, contentTypeSlug string, documentIDs []string, locale string) ([]*entity.Document, error) {
	q := r.table(contentTypeSlug).WithContext(ctx).
		Where("version = ? AND locale = ? AND document_id IN ?",
			entity.VersionPublished, locale, documentIDs)
	return r.findMany(ctx, contentTypeSlug, q)
}

func (r *documentRepository) DeleteByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND locale = ?", documentID, locale).
		Delete(map[string]any{}).Error
}

func (r *documentRepository) DeletePublishedByDocumentID(ctx context.Context, contentTypeSlug, documentID, locale string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Where("document_id = ? AND version = ? AND locale = ?",
			documentID, entity.VersionPublished, locale).
		Delete(map[string]any{}).Error
}

func (r *documentRepository) DeleteAllByContentType(ctx context.Context, contentTypeSlug string) error {
	return r.table(contentTypeSlug).WithContext(ctx).
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(map[string]any{}).Error
}

// helpers

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toUint(v any) uint {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int32:
		return uint(val)
	case int64:
		return uint(val)
	case uint:
		return val
	case float64:
		return uint(val)
	case int:
		return uint(val)
	default:
		return 0
	}
}

func toTime(v any) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		t, _ := time.Parse(time.RFC3339Nano, val)
		return t
	default:
		return time.Time{}
	}
}

