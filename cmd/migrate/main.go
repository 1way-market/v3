package main

import (
	"database/sql"
	"fmt"
	"github.com/1way-market/v3/internal/config"
	_ "github.com/lib/pq"
	"log"
)

func main() {

	cfg := config.New()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}

func runMigrations(db *sql.DB) error {
	// Create a temporary column to store the numeric status
	if _, err := db.Exec(`ALTER TABLE ads ADD COLUMN status_new INTEGER`); err != nil {
		return fmt.Errorf("failed to add status_new column: %v", err)
	}

	// Convert existing status values to integers
	if _, err := db.Exec(`
		UPDATE ads 
		SET status_new = CASE status
			WHEN 'draft' THEN 0
			WHEN 'pending' THEN 1
			WHEN 'from_parser' THEN 2
			WHEN 'active' THEN 3
			WHEN 'completed' THEN 4
			WHEN 'rejected' THEN 5
			WHEN 'approved' THEN 6
			WHEN 'unknown' THEN 7
			WHEN 'duplicate' THEN 8
			ELSE 0
		END
	`); err != nil {
		return fmt.Errorf("failed to update status values: %v", err)
	}

	// Drop the old status column and rename the new one
	if _, err := db.Exec(`
		ALTER TABLE ads 
		DROP COLUMN status,
		ALTER COLUMN status_new SET NOT NULL,
		ALTER COLUMN status_new SET DEFAULT 0,
		RENAME COLUMN status_new TO status
	`); err != nil {
		return fmt.Errorf("failed to rename status column: %v", err)
	}

	// Recreate the status index
	if _, err := db.Exec(`CREATE INDEX idx_ads_status ON ads(status)`); err != nil {
		return fmt.Errorf("failed to create status index: %v", err)
	}

	// Create properties table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS properties (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			value_type VARCHAR(50) NOT NULL,
			is_searchable BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create properties table: %v", err)
	}

	// Create property_values table
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS property_values (
			id SERIAL PRIMARY KEY,
			property_id INTEGER NOT NULL REFERENCES properties(id),
			value TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("failed to create property_values table: %v", err)
	}

	// Create indexes
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_property_values_property_id ON property_values(property_id);
		CREATE INDEX IF NOT EXISTS idx_properties_type ON properties(type);
		CREATE INDEX IF NOT EXISTS idx_properties_searchable ON properties(is_searchable) WHERE is_searchable = true;
	`); err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	// Update ads table to use jsonb for properties
	if _, err := db.Exec(`
		ALTER TABLE ads 
		DROP COLUMN IF EXISTS properties,
		ADD COLUMN IF NOT EXISTS properties JSONB;
		CREATE INDEX IF NOT EXISTS idx_ads_properties ON ads USING gin(properties);
	`); err != nil {
		return fmt.Errorf("failed to update ads table: %v", err)
	}

	return nil
}
