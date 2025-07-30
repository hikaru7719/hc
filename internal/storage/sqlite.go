package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/hc/hc/internal/logger"
	"github.com/hc/hc/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// SQL queries as constants
const (
	// Table creation queries
	createFoldersTableQuery = `
		CREATE TABLE IF NOT EXISTS folders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
		)`

	createRequestsTableQuery = `
		CREATE TABLE IF NOT EXISTS requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			folder_id INTEGER,
			method TEXT NOT NULL,
			url TEXT NOT NULL,
			headers TEXT,
			body TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE SET NULL
		)`

	// Folder queries
	insertFolderQuery  = `INSERT INTO folders (name, parent_id) VALUES (?, ?)`
	selectFolderQuery  = `SELECT id, name, parent_id, created_at, updated_at FROM folders WHERE id = ?`
	selectFoldersQuery = `SELECT id, name, parent_id, created_at, updated_at FROM folders ORDER BY name`
	updateFolderQuery  = `UPDATE folders SET name = ?, parent_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	deleteFolderQuery  = `DELETE FROM folders WHERE id = ?`

	// Request queries
	insertRequestQuery  = `INSERT INTO requests (name, folder_id, method, url, headers, body) VALUES (?, ?, ?, ?, ?, ?)`
	selectRequestQuery  = `SELECT id, name, folder_id, method, url, headers, body, created_at, updated_at FROM requests WHERE id = ?`
	selectRequestsQuery = `SELECT id, name, folder_id, method, url, headers, body, created_at, updated_at FROM requests ORDER BY updated_at DESC`
	updateRequestQuery  = `UPDATE requests SET name = ?, folder_id = ?, method = ?, url = ?, headers = ?, body = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	deleteRequestQuery  = `DELETE FROM requests WHERE id = ?`
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
	log *slog.Logger
}

// InitDB initializes the database connection and creates tables
func InitDB() (*DB, error) {
	log := logger.Get()

	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	log.Info("Opening database", slog.String("path", dbPath))

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	wrapper := &DB{
		DB:  db,
		log: log,
	}

	if err := wrapper.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Info("Database initialized successfully")
	return wrapper, nil
}

// getDBPath is a variable holding the function to get database path (for testing)
var getDBPath = defaultGetDBPath

// defaultGetDBPath returns the path to the database file
func defaultGetDBPath() (string, error) {
	// Check for test database path
	if testPath := os.Getenv("HC_TEST_DB_PATH"); testPath != "" {
		return testPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dbDir := filepath.Join(homeDir, ".hc")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dbDir, "hc.db"), nil
}

// createTables creates the necessary database tables
func (db *DB) createTables() error {
	queries := []string{
		createFoldersTableQuery,
		createRequestsTableQuery,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			db.log.Error("Failed to rollback transaction", slog.String("error", rbErr.Error()))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Folder operations

// CreateFolder creates a new folder in the database
func (db *DB) CreateFolder(folder *models.Folder) error {
	db.log.Info("Creating folder", slog.String("name", folder.Name))

	result, err := db.Exec(insertFolderQuery, folder.Name, folder.ParentID)
	if err != nil {
		db.log.Error("Failed to create folder", slog.String("error", err.Error()))
		return fmt.Errorf("failed to create folder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	folder.ID = int(id)
	return db.GetFolder(folder.ID, folder)
}

// GetFolder retrieves a folder by ID
func (db *DB) GetFolder(id int, folder *models.Folder) error {
	err := db.QueryRow(selectFolderQuery, id).Scan(
		&folder.ID,
		&folder.Name,
		&folder.ParentID,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return fmt.Errorf("folder not found")
	}
	if err != nil {
		db.log.Error("Failed to get folder", slog.Int("id", id), slog.String("error", err.Error()))
		return fmt.Errorf("failed to get folder: %w", err)
	}

	return nil
}

// GetFolders retrieves all folders
func (db *DB) GetFolders() ([]models.Folder, error) {
	rows, err := db.Query(selectFoldersQuery)
	if err != nil {
		db.log.Error("Failed to get folders", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get folders: %w", err)
	}
	defer rows.Close()

	folders := make([]models.Folder, 0)
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.Name,
			&folder.ParentID,
			&folder.CreatedAt,
			&folder.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, folder)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating folders: %w", err)
	}

	return folders, nil
}

// UpdateFolder updates an existing folder
func (db *DB) UpdateFolder(folder *models.Folder) error {
	db.log.Info("Updating folder", slog.Int("id", folder.ID))

	result, err := db.Exec(updateFolderQuery, folder.Name, folder.ParentID, folder.ID)
	if err != nil {
		db.log.Error("Failed to update folder", slog.String("error", err.Error()))
		return fmt.Errorf("failed to update folder: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("folder not found")
	}

	return nil
}

// DeleteFolder deletes a folder by ID
func (db *DB) DeleteFolder(id int) error {
	db.log.Info("Deleting folder", slog.Int("id", id))

	result, err := db.Exec(deleteFolderQuery, id)
	if err != nil {
		db.log.Error("Failed to delete folder", slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("folder not found")
	}

	return nil
}

// Request operations

// CreateRequest creates a new request in the database
func (db *DB) CreateRequest(request *models.Request) error {
	db.log.Info("Creating request", slog.String("name", request.Name))

	headersJSON, err := serializeHeaders(request.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize headers: %w", err)
	}

	result, err := db.Exec(insertRequestQuery,
		request.Name,
		request.FolderID,
		request.Method,
		request.URL,
		headersJSON,
		request.Body,
	)
	if err != nil {
		db.log.Error("Failed to create request", slog.String("error", err.Error()))
		return fmt.Errorf("failed to create request: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	request.ID = int(id)
	return db.GetRequest(request.ID, request)
}

// GetRequest retrieves a request by ID
func (db *DB) GetRequest(id int, request *models.Request) error {
	var headersStr string
	err := db.QueryRow(selectRequestQuery, id).Scan(
		&request.ID,
		&request.Name,
		&request.FolderID,
		&request.Method,
		&request.URL,
		&headersStr,
		&request.Body,
		&request.CreatedAt,
		&request.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return fmt.Errorf("request not found")
	}
	if err != nil {
		db.log.Error("Failed to get request", slog.Int("id", id), slog.String("error", err.Error()))
		return fmt.Errorf("failed to get request: %w", err)
	}

	headers, err := deserializeHeaders(headersStr)
	if err != nil {
		return fmt.Errorf("failed to deserialize headers: %w", err)
	}
	request.Headers = headers

	return nil
}

// GetRequests retrieves all requests
func (db *DB) GetRequests() ([]models.Request, error) {
	rows, err := db.Query(selectRequestsQuery)
	if err != nil {
		db.log.Error("Failed to get requests", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get requests: %w", err)
	}
	defer rows.Close()

	requests := make([]models.Request, 0)
	for rows.Next() {
		var request models.Request
		var headersStr string

		if err := rows.Scan(
			&request.ID,
			&request.Name,
			&request.FolderID,
			&request.Method,
			&request.URL,
			&headersStr,
			&request.Body,
			&request.CreatedAt,
			&request.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan request: %w", err)
		}

		headers, err := deserializeHeaders(headersStr)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize headers: %w", err)
		}
		request.Headers = headers

		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating requests: %w", err)
	}

	return requests, nil
}

// UpdateRequest updates an existing request
func (db *DB) UpdateRequest(request *models.Request) error {
	db.log.Info("Updating request", slog.Int("id", request.ID))

	headersJSON, err := serializeHeaders(request.Headers)
	if err != nil {
		return fmt.Errorf("failed to serialize headers: %w", err)
	}

	result, err := db.Exec(updateRequestQuery,
		request.Name,
		request.FolderID,
		request.Method,
		request.URL,
		headersJSON,
		request.Body,
		request.ID,
	)
	if err != nil {
		db.log.Error("Failed to update request", slog.String("error", err.Error()))
		return fmt.Errorf("failed to update request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("request not found")
	}

	return nil
}

// DeleteRequest deletes a request by ID
func (db *DB) DeleteRequest(id int) error {
	db.log.Info("Deleting request", slog.Int("id", id))

	result, err := db.Exec(deleteRequestQuery, id)
	if err != nil {
		db.log.Error("Failed to delete request", slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete request: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("request not found")
	}

	return nil
}

// Helper functions

// serializeHeaders converts headers map to JSON string
func serializeHeaders(headers map[string]string) (string, error) {
	if headers == nil {
		headers = make(map[string]string)
	}

	data, err := json.Marshal(headers)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// deserializeHeaders converts JSON string to headers map
func deserializeHeaders(headersStr string) (map[string]string, error) {
	headers := make(map[string]string)

	if headersStr == "" {
		return headers, nil
	}

	if err := json.Unmarshal([]byte(headersStr), &headers); err != nil {
		return nil, err
	}

	return headers, nil
}
