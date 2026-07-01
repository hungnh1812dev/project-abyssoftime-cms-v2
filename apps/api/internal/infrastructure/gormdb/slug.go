package gormdb

import "strings"

func sanitizeSlug(slug string) string {
	return toSnakeCase(strings.ReplaceAll(slug, "-", "_"))
}

func documentTableName(slug string) string {
	return "documents_" + sanitizeSlug(slug)
}

func componentTableName(slug, componentName string) string {
	return "components_" + sanitizeSlug(slug) + "_" + sanitizeSlug(componentName)
}
