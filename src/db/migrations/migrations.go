package migrations

import (
	"embed"
	"io/fs"
	"strings"
)

type Migration struct {
	Version string
	Script  string
}

//go:embed *.sql
var files embed.FS

// GetAllMigrations returns parsed Migration structs sorted by filename
func GetAllMigrations() []Migration {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil
	}

	var m []Migration
	for _, entry := range entries {
		if !entry.IsDir() {
			content, _ := fs.ReadFile(files, entry.Name())
			name := entry.Name()

			// Remove .sql extension
			if strings.HasSuffix(name, ".sql") {
				name = strings.TrimSuffix(name, ".sql")
			}

			// Split by underscore and take the second part
			parts := strings.Split(name, "_")
			if len(parts) > 1 {
				name = parts[1]
			}

			m = append(m, Migration{
				Version: name,
				Script:  string(content),
			})
		}
	}
	return m
}
