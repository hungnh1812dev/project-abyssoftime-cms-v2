package content_type

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type ContentTypeDefinition struct {
	Slug       string                   `json:"slug"`
	Name       string                   `json:"name"`
	Kind       string                   `json:"kind"`
	ListFields []string                 `json:"listFields,omitempty"`
	Fields     []entity.FieldDefinition `json:"fields"`
}

// LoadDefinitions reads every *.json file in dir and parses it into a
// ContentTypeDefinition. It is the source of truth read by Sync on startup.
func LoadDefinitions(dir string) ([]ContentTypeDefinition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read content-type definitions dir %q: %w", dir, err)
	}

	var defs []ContentTypeDefinition
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", path, err)
		}

		var def ContentTypeDefinition
		if err := json.Unmarshal(data, &def); err != nil {
			return nil, fmt.Errorf("parse %q: %w", path, err)
		}
		if err := validateDefinition(def, path); err != nil {
			return nil, err
		}
		defs = append(defs, def)
	}

	return defs, nil
}

func validateDefinition(def ContentTypeDefinition, path string) error {
	if err := validateFields(def.Fields, path, 1); err != nil {
		return err
	}
	return validateListFields(def, path)
}

func validateListFields(def ContentTypeDefinition, path string) error {
	if len(def.ListFields) == 0 {
		return nil
	}
	fieldNames := make(map[string]bool, len(def.Fields))
	for _, f := range def.Fields {
		fieldNames[f.Name] = true
	}
	for _, lf := range def.ListFields {
		if !fieldNames[lf] {
			return fmt.Errorf("%q: listFields entry %q does not match any field name", path, lf)
		}
	}
	return nil
}

func validateFields(fields []entity.FieldDefinition, path string, depth int) error {
	for _, f := range fields {
		switch f.Type {
		case "layout":
			if len(f.Fields) == 0 {
				return fmt.Errorf("%q: layout field %q must have at least one child field", path, f.Name)
			}
			for _, child := range f.Fields {
				if child.Type == "component" {
					return fmt.Errorf("%q: layout field %q must not contain component children", path, f.Name)
				}
			}
		case "component":
			if f.Name == "" {
				return fmt.Errorf("%q: component field must have a non-empty name", path)
			}
			if depth > 1 {
				return fmt.Errorf("%q: component %q exceeds maximum nesting depth of 2", path, f.Name)
			}
			if err := validateFields(f.Fields, path, depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}
