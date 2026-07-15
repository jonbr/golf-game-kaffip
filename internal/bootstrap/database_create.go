package bootstrap

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

type migrateLogger struct{}

func ensureDatabaseExists(dsn string) error {
	// Parse the DATABASE_URL
	u, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// Extract the target DB name
	targetDB := strings.TrimPrefix(u.Path, "/")

	// Replace DB name with "postgres" for admin connection
	u.Path = "/postgres"
	adminDSN := u.String()

	// Connect to the default "postgres" database
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	defer adminDB.Close()

	// Check if the target DB exists
	var exists bool
	err = adminDB.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM pg_database WHERE datname = $1
        );
    `, targetDB).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed checking database existence: %w", err)
	}

	if exists {
		return nil
	}

	// Create the database
	_, err = adminDB.Exec(`CREATE DATABASE golf_game;`)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	fmt.Println("Database golf_game created successfully")
	return nil
}

func (m migrateLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (m migrateLogger) Verbose() bool {
	return true
}
