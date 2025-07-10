package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database connection pool.
var DB *sql.DB

// InitDB initializes the database connection and applies migrations.
func InitDB() (*sql.DB, error) {
	dbPath := "./database/social_network.db"
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err = DB.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database initialized successfully.")
	applyMigrations() // We don't need to pass the dbPath anymore

	return DB, nil
}

func applyMigrations() {
	log.Println("Applying database migrations...")


	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "migrations")
	// The path must be in file:// format for the migrate library
	migrationsURL := "file://" + migrationsPath


	driver, err := sqlite3.WithInstance(DB, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("could not create sqlite3 driver instance: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL, 
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("failed to get migration version: %v", err)
	}
	if err == migrate.ErrNilVersion {
		log.Println("No migrations found or applied. This might be an error.")
	} else {
		log.Printf("Current migration version: %d, Dirty: %v\n", version, dirty)
	}

	log.Println("Database migrations applied successfully.")
}