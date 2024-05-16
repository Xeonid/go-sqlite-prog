package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"go-sqlite-prog/bugz"
	"log"
	"os"
	"path/filepath"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed schema.sql
var schemaFS embed.FS

func main() {
	db, err := sqlite.OpenConn("bugs.db", 0)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Read the schema from the embedded file
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		log.Fatalf("failed to read schema: %w", err)
	}

	// Execute the schema
	if err := sqlitex.ExecuteTransient(db, string(schema), nil); err != nil {
		log.Fatalf("failed to execute schema: %w", err)
	}

	err = importBugsFromJSON(db, ".")
	if err != nil {
		log.Fatalf("Error importing bugs from JSON: %v", err)
	}

	fmt.Println("Database and schema are ready. Bugs imported successfully.")
}

func importBugsFromJSON(db *sqlite.Conn, directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(directory, file.Name())
		bug := bugz.Bug{}
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		if err := json.Unmarshal(fileData, &bug); err != nil {
			return fmt.Errorf("error decoding JSON from file %s: %v", filePath, err)
		}

		bugJSON, err := json.Marshal(bug)
		if err != nil {
			return fmt.Errorf("error marshaling Bug to JSON: %v", err)
		}

		// Execute the INSERT query
		if err := sqlitex.Exec(db, "INSERT INTO bugs (id, CreationTime, Creator, Summary, OtherFieldsJSON) VALUES (?, ?, ?, ?, ?)", nil, bug.ID, bug.CreationTime, bug.Creator, bug.Summary, string(bugJSON)); err != nil {
			return fmt.Errorf("error executing statement: %v", err)
		}
	}

	return nil
}
