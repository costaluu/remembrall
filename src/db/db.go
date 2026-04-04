package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "embed"

	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/db/migrations"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func Open() (*sql.DB, error) {
	config := config.LoadConfig()
	dbLocation := config.DatabaseLocation

	driver := "sqlite" // Padrão local

	// Se a URL começar com libsql ou http, troca o driver para o da Turso
	if strings.HasPrefix(dbLocation, "libsql://") || strings.HasPrefix(dbLocation, "https://") {
		driver = "libsql"
	} else {
		dbLocation = utils.ReplaceTildeWithHomeDir(dbLocation)
	}

	db, err := sql.Open(driver, dbLocation)

	if err != nil {
		logger.Fatal(fmt.Errorf("erro ao configurar driver: %w", err))
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func ApplyAllMigrations() {
	db, err := Open()

	if err != nil {
		logger.Fatal(fmt.Errorf("failed to open database: %w", err))
	}

	migrations := migrations.GetAllMigrations()

	for _, migration := range migrations {
		_, err := db.Exec(migration.Script)
		if err != nil {
			logger.Fatal(fmt.Errorf("failed to execute migration: %w", err))
		}
	}
}

func ApplyMigrations() {
	db, err := Open()

	if err != nil {
		logger.Fatal(fmt.Errorf("failed to open database: %w", err))
	}

	migrations := migrations.GetAllMigrations()

	for _, migration := range migrations {
		var exists bool

		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", migration.Version).Scan(&exists)

		if err != nil {
			logger.Fatal(fmt.Errorf("failed to check migration status: %w", err))
		}

		if !exists {
			_, err := db.Exec(migration.Script)

			if err != nil {
				logger.Fatal(fmt.Errorf("failed to execute migration: %w", err))
			}

			_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)

			if err != nil {
				logger.Fatal(fmt.Errorf("failed to record migration: %w", err))
			}
		}
	}
}
