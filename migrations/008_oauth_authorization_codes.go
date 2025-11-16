package main

import (
	"database/sql"
	"fmt"
	"os"
)

func main() {
	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var sqlFile string
	if direction == "down" {
		sqlFile = "migrations/008_oauth_authorization_codes.down.sql"
	} else {
		sqlFile = "migrations/008_oauth_authorization_codes.up.sql"
	}

	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		fmt.Printf("Failed to read SQL file: %v\n", err)
		os.Exit(1)
	}

	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		fmt.Printf("Failed to execute migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Migration 008_oauth_authorization_codes %s completed successfully\n", direction)
}
