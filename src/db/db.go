package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"

	"github.com/costaluu/taskthing/src/config"
	"github.com/costaluu/taskthing/src/db/migrations"
	"github.com/costaluu/taskthing/src/logger"
	"github.com/costaluu/taskthing/src/utils"
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

		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)

		if err != nil {
			logger.Fatal(fmt.Errorf("failed to record migration: %w", err))
		}
	}
}

func ApplyMigrations() bool {
	var appliedAnyMigration bool = false

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
			appliedAnyMigration = true

			logger.Info(fmt.Sprintf("applying migration %s...", migration.Version))

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

	return appliedAnyMigration
}

func SchemaMigrationsExists(db *sql.DB) bool {
	var exists bool
	var tableName string = "schema_migrations"

	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`

	err := db.QueryRow(query, tableName).Scan(&exists)

	if err != nil {
		logger.Fatal(err)
	}

	return exists
}

func TruncateTasks(db *sql.DB) {
	if _, err := db.Exec("DELETE FROM tasks"); err != nil {
		logger.Fatal(fmt.Errorf("truncate tasks: %w", err))
	}
}

// Migrate copia todos os dados de src para dst.
// O schema é recriado via ApplyAllMigrations antes da cópia.
func MigrateDatabases(src, dst *sql.DB) {
	var migrateTables = []string{"schema_migrations", "tasks"}

	// Pragmas para acelerar inserções em massa no destino
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=OFF;",
		"PRAGMA foreign_keys=OFF;",
	} {
		if _, err := dst.Exec(pragma); err != nil {
			logger.Fatal(fmt.Errorf("migrate: pragma %q: %w", pragma, err))
		}
	}

	for _, table := range migrateTables {
		n, err := migrateTable(src, dst, table)

		if err != nil {
			logger.Fatal(fmt.Errorf("migrate: table %q: %w", table, err))
		}

		logger.Info(fmt.Sprintf("migrate: %q — %d linha(s) copiada(s)", table, n))
	}
}

func migrateTable(src, dst *sql.DB, table string) (int, error) {
	const migrateChunkSize = 500

	rows, err := src.Query(fmt.Sprintf("SELECT * FROM %s", table))
	if err != nil {
		return 0, fmt.Errorf("select: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return 0, fmt.Errorf("columns: %w", err)
	}

	placeholders := strings.Repeat("?,", len(cols))
	insertSQL := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(cols, ","),
		placeholders[:len(placeholders)-1],
	)

	// Ponteiros reutilizáveis para o scan
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	var (
		tx    *sql.Tx
		stmt  *sql.Stmt
		total int
	)

	begin := func() error {
		tx, err = dst.Begin()
		if err != nil {
			return fmt.Errorf("begin: %w", err)
		}
		stmt, err = tx.Prepare(insertSQL)
		return err
	}

	commit := func() error {
		stmt.Close()
		return tx.Commit()
	}

	if err := begin(); err != nil {
		return 0, err
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return total, fmt.Errorf("scan: %w", err)
		}
		if _, err := stmt.Exec(vals...); err != nil {
			return total, fmt.Errorf("insert: %w", err)
		}
		total++

		if total%migrateChunkSize == 0 {
			if err := commit(); err != nil {
				return total, err
			}
			if err := begin(); err != nil {
				return total, err
			}
		}
	}

	if err := rows.Err(); err != nil {
		return total, err
	}

	return total, commit()
}
