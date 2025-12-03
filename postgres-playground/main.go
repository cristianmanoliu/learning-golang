package main

import (
	// Database/sql is the standard library package for SQL database access
	"database/sql"
	// Fmt and log are standard library packages for formatting and logging
	"fmt"
	// Log package for logging errors
	"log"
	// Lib/pq is the PostgreSQL driver for database/sql
	_ "github.com/lib/pq"
)

func main() {
	// Connection string must match docker-compose settings
	dsn := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

	// Open DB connection (does not actually connect yet)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Verify the DB is reachable
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		log.Fatalf("failed to run select 1: %v", err)
	}

	fmt.Println("SELECT 1 result:", result)
}
