package main

import (
	// Database/sql is the generic Go database package.
	"database/sql"
	// Fmt is used for printing.
	"fmt"
)

// User is a simple model for the users table.
// In Java this would be a POJO.
type User struct {
	ID   int
	Name string
}

// runUserDemo creates the table (if needed), inserts a user,
// then reads and prints all users.
func runUserDemo(db *sql.DB) error {
	// Ensure the users table exists.
	if err := ensureUsersTable(db); err != nil {
		return fmt.Errorf("ensure users table: %w", err)
	}

	// Insert a user.
	if err := insertUser(db, "Cristi"); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	// Get all users.
	users, err := getAllUsers(db)
	if err != nil {
		return fmt.Errorf("get all users: %w", err)
	}

	fmt.Println("Users in DB:")
	// Print all users.
	for _, u := range users {
		fmt.Printf("  id=%d, name=%s\n", u.ID, u.Name)
	}

	// Return no error.
	return nil
}

func ensureUsersTable(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS users (
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);`
	_, err := db.Exec(ddl)
	return err
}

func insertUser(db *sql.DB, name string) error {
	const insertSQL = `INSERT INTO users (name) VALUES ($1);`
	_, err := db.Exec(insertSQL, name)
	return err
}

func getAllUsers(db *sql.DB) ([]User, error) {
	const query = `SELECT id, name FROM users ORDER BY id;`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return the result slice and no error.
	return result, nil
}
