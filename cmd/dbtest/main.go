package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/parsel-email/mailroom/internal/database"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_FILE")
	if dbPath == "" {
		dbPath = "./db.sqlite"
	}

	fmt.Printf("Testing SQLite database at %s\n", dbPath)

	// Initialize the database
	db, err := database.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run extension tests
	fmt.Println("\n=== Testing SQLite Extensions ===")
	runExtensionTests(db)
}

func runExtensionTests(db *database.SQLite) {
	// Test FTS5
	fmt.Println("\n1. Testing FTS5 Extension:")
	testFTS5(db.GetDB())

	// Test JSON Functions
	fmt.Println("\n2. Testing JSON Functions:")
	testJSON(db.GetDB())

	fmt.Println("\n✅ All extension tests completed successfully!")
}

func testFTS5(db *sql.DB) {
	// Create a temporary FTS5 table
	_, err := db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS temp_fts USING fts5(content);
	`)
	if err != nil {
		log.Fatalf("Failed to create FTS5 table: %v", err)
	}
	defer db.Exec("DROP TABLE IF EXISTS temp_fts")

	// Insert some test data
	_, err = db.Exec(`
		INSERT INTO temp_fts (content) VALUES 
		('The quick brown fox jumps over the lazy dog'),
		('Lorem ipsum dolor sit amet, consectetur adipiscing elit'),
		('SQLite FTS5 is working properly');
	`)
	if err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}

	// Test a simple FTS5 query
	var count int
	err = db.QueryRow(`SELECT count(*) FROM temp_fts WHERE temp_fts MATCH 'quick'`).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to query FTS5 table: %v", err)
	}

	if count != 1 {
		log.Fatalf("FTS5 query returned unexpected result: %d (expected 1)", count)
	}

	fmt.Println("✓ FTS5 is working correctly")
}

func testJSON(db *sql.DB) {
	// Test JSON functions
	var result string
	err := db.QueryRow(`SELECT json_extract('{"name": "SQLite", "features": ["FTS5", "JSON"]}', '$.name')`).Scan(&result)
	if err != nil {
		log.Fatalf("Failed to execute JSON function: %v", err)
	}

	if result != "SQLite" {
		log.Fatalf("JSON function returned unexpected result: %s (expected 'SQLite')", result)
	}

	// Test a more complex JSON query
	var count int
	err = db.QueryRow(`
		WITH json_data AS (
			SELECT '{"extensions": ["FTS5", "JSON", "RTREE"]}' AS json_doc
		)
		SELECT json_array_length(json_extract(json_doc, '$.extensions')) 
		FROM json_data
	`).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to execute complex JSON query: %v", err)
	}

	if count != 3 {
		log.Fatalf("JSON array length function returned unexpected result: %d (expected 3)", count)
	}

	fmt.Println("✓ JSON functions are working correctly")
}
