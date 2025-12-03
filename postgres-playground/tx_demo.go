package main

import (
	"database/sql"
	"fmt"
)

// runTransactionDemo shows how to insert multiple users in a single transaction.
func runTransactionDemo(db *sql.DB) error {
	// Start a new transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// If something goes wrong, we roll back.
	// If everything is fine, we'll Commit() at the end.
	const insertSQL = `INSERT INTO users (name) VALUES ($1);`

	if _, err := tx.Exec(insertSQL, "TxUser-1"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert first user in tx: %w", err)
	}

	if _, err := tx.Exec(insertSQL, "TxUser-2"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert second user in tx: %w", err)
	}

	// All good → commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	fmt.Println("Inserted TxUser-1 and TxUser-2 in a single transaction")

	return nil
}
