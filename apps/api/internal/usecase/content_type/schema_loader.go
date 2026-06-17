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
	Slug   string                   `json:"slug"`
	Name   string                   `json:"name"`
	Kind   string                   `json:"kind"`
	Fields []entity.FieldDefinition `json:"fields"`
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
		defs = append(defs, def)
	}

	return defs, nil
}
