package database

import (
	"database/sql"
	"fmt"
	"strings"
)

type ColumnInfo struct {
	Name          string
	DataType      string
	IsNullable    string
	ColumnDefault *string
	IsSerial      bool
}

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
	Indexes []string
}

func ValidateSchema(db *sql.DB) error {
	// Expected schema definition
	expectedTables := map[string]TableInfo{
		"ads": {
			Name: "ads",
			Columns: []ColumnInfo{
				{"id", "integer", "NO", nil, true}, // Serial/auto-increment column
				{"title", "jsonb", "NO", nil, false},
				{"description", "jsonb", "YES", nil, false},
				{"properties", "jsonb", "YES", nil, false},
				{"category_ids", "ARRAY", "YES", nil, false}, // Changed to match PostgreSQL's type
				{"status", "integer", "NO", strPtr("0"), false},
				{"price", "jsonb", "YES", nil, false},
				{"search_vector", "tsvector", "YES", nil, false},
				{"created_at", "timestamp with time zone", "YES", strPtr("CURRENT_TIMESTAMP"), false},
				{"updated_at", "timestamp with time zone", "YES", strPtr("CURRENT_TIMESTAMP"), false},
			},
			Indexes: []string{
				"ads_pkey",
				"idx_ads_status",
				"idx_ads_category_ids",
				"idx_ads_search_vector",
				"idx_ads_title",
				"idx_ads_properties",
				"idx_ads_price",
				"idx_ads_created_at",
			},
		},
		"category_closure": {
			Name: "category_closure",
			Columns: []ColumnInfo{
				{"ancestor_id", "integer", "NO", nil, false},
				{"descendant_id", "integer", "NO", nil, false},
				{"depth", "integer", "NO", nil, false},
			},
			Indexes: []string{
				"category_closure_pkey",
				"idx_category_closure_ancestor",
				"idx_category_closure_descendant",
			},
		},
	}

	// Check each expected table
	for tableName, expectedTable := range expectedTables {
		// Check if table exists
		if !tableExists(db, tableName) {
			return fmt.Errorf("table %s does not exist", tableName)
		}

		// Get actual columns
		actualColumns, err := getTableColumns(db, tableName)
		if err != nil {
			return fmt.Errorf("error getting columns for table %s: %v", tableName, err)
		}

		// Compare columns
		for _, expectedCol := range expectedTable.Columns {
			found := false
			for _, actualCol := range actualColumns {
				if expectedCol.Name == actualCol.Name {
					found = true
					if err := compareColumns(expectedCol, actualCol); err != nil {
						return fmt.Errorf("column mismatch in table %s: %v", tableName, err)
					}
					break
				}
			}
			if !found {
				return fmt.Errorf("missing column %s in table %s", expectedCol.Name, tableName)
			}
		}

		// Check for extra columns
		for _, actualCol := range actualColumns {
			found := false
			for _, expectedCol := range expectedTable.Columns {
				if actualCol.Name == expectedCol.Name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("extra column %s found in table %s", actualCol.Name, tableName)
			}
		}

		// Check indexes
		actualIndexes, err := getTableIndexes(db, tableName)
		if err != nil {
			return fmt.Errorf("error getting indexes for table %s: %v", tableName, err)
		}

		for _, expectedIdx := range expectedTable.Indexes {
			found := false
			for _, actualIdx := range actualIndexes {
				if strings.EqualFold(expectedIdx, actualIdx) {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("missing index %s in table %s", expectedIdx, tableName)
			}
		}
	}

	return nil
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename = $1
	)`
	db.QueryRow(query, tableName).Scan(&exists)
	return exists
}

func getTableColumns(db *sql.DB, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			c.column_name,
			CASE 
				WHEN c.data_type = 'ARRAY' THEN 'ARRAY'
				WHEN c.data_type = 'USER-DEFINED' THEN c.udt_name
				ELSE c.data_type
			END as data_type,
			c.is_nullable,
			c.column_default,
			COALESCE(
				(EXISTS (
					SELECT 1 FROM pg_attribute a
					JOIN pg_class t ON a.attrelid = t.oid
					JOIN pg_namespace n ON t.relnamespace = n.oid
					WHERE n.nspname = 'public'
					AND t.relname = $1
					AND a.attname = c.column_name
					AND a.attidentity = 'a'
				) OR c.column_default LIKE 'nextval%'
				), false) as is_serial
		FROM information_schema.columns c
		WHERE c.table_schema = 'public'
		AND c.table_name = $1
		ORDER BY c.ordinal_position`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable, &col.ColumnDefault, &col.IsSerial); err != nil {
			return nil, fmt.Errorf("error scanning column info: %v", err)
		}
		columns = append(columns, col)
	}
	return columns, nil
}

func getTableIndexes(db *sql.DB, tableName string) ([]string, error) {
	query := `
		SELECT indexname 
		FROM pg_indexes 
		WHERE schemaname = 'public' 
		AND tablename = $1`

	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []string
	for rows.Next() {
		var indexName string
		if err := rows.Scan(&indexName); err != nil {
			return nil, err
		}
		indexes = append(indexes, indexName)
	}
	return indexes, nil
}

func compareColumns(expected, actual ColumnInfo) error {
	// Normalize data types for comparison
	expectedType := normalizeDataType(expected.DataType)
	actualType := normalizeDataType(actual.DataType)

	if expectedType != actualType {
		return fmt.Errorf("column %s: expected type %s, got %s",
			expected.Name, expectedType, actualType)
	}
	if expected.IsNullable != actual.IsNullable {
		return fmt.Errorf("column %s: expected nullable %s, got %s",
			expected.Name, expected.IsNullable, actual.IsNullable)
	}

	// For serial columns, we don't compare the default value
	if expected.IsSerial {
		if !actual.IsSerial {
			return fmt.Errorf("column %s: expected serial/identity column", expected.Name)
		}
		return nil
	}

	// For non-serial columns, compare default values
	if (expected.ColumnDefault == nil && actual.ColumnDefault != nil) ||
		(expected.ColumnDefault != nil && actual.ColumnDefault == nil) ||
		(expected.ColumnDefault != nil && actual.ColumnDefault != nil &&
			!strings.Contains(*actual.ColumnDefault, *expected.ColumnDefault)) {
		return fmt.Errorf("column %s: default value mismatch", expected.Name)
	}
	return nil
}

func normalizeDataType(dataType string) string {
	// Convert data type to uppercase for consistent comparison
	dataType = strings.ToUpper(dataType)

	// Handle array types
	if strings.HasPrefix(dataType, "_") || strings.HasSuffix(dataType, "[]") || dataType == "ARRAY" {
		return "ARRAY"
	}

	// Handle user-defined types
	switch dataType {
	case "JSONB", "JSON":
		return "JSONB"
	case "TSVECTOR":
		return "TSVECTOR"
	case "CHARACTER VARYING", "VARCHAR":
		return "CHARACTER VARYING"
	case "INTEGER", "INT", "INT4":
		return "INTEGER"
	case "NUMERIC", "DECIMAL":
		return "NUMERIC"
	case "TIMESTAMP WITH TIME ZONE", "TIMESTAMPTZ":
		return "TIMESTAMP WITH TIME ZONE"
	default:
		return dataType
	}
}

func strPtr(s string) *string {
	return &s
}
