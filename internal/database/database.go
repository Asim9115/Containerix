package database

import (
    "database/sql"
    "embed"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    _ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var db *sql.DB

func Init(dbPath string) error {
    dir := filepath.Dir(dbPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create db directory: %w", err)
    }

    var err error
    db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
    if err != nil {
        return fmt.Errorf("open database: %w", err)
    }

    // SQLite performance tuning
    pragmas := []string{
        "PRAGMA journal_mode=WAL",
        "PRAGMA synchronous=NORMAL",
        "PRAGMA foreign_keys=ON",
        "PRAGMA busy_timeout=5000",
    }
    for _, p := range pragmas {
        if _, err := db.Exec(p); err != nil {
            return fmt.Errorf("exec %s: %w", p, err)
        }
    }

    if err := runMigrations(); err != nil {
        return fmt.Errorf("migrations: %w", err)
    }

    log.Println("Database initialized:", dbPath)
    return nil
}

func GetDB() *sql.DB { return db }

func Close() error {
    if db != nil {
        return db.Close()
    }
    return nil
}

func runMigrations() error {
    sourceDriver, err := iofs.New(migrationsFS, "migrations")
    if err != nil {
        return fmt.Errorf("create source driver: %w", err)
    }

    dbDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
    if err != nil {
        return fmt.Errorf("create db driver: %w", err)
    }

    m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
    if err != nil {
        return fmt.Errorf("create migrate instance: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("run migrations: %w", err)
    }

    version, dirty, _ := m.Version()
    log.Printf("migrations: at version %d (dirty=%v)", version, dirty)

    return nil
}