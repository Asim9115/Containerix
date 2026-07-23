package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    _"github.com/mattn/go-sqlite3"
)

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
    schema := `...` // embed 001_init.sql content here or use embed
    _, err := db.Exec(schema)
    return err
}
