package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hc/hc/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func InitDB() (*DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".hc")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	dbPath := filepath.Join(dbDir, "hc.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &DB{db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS folders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS requests (
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
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) CreateFolder(folder *models.Folder) error {
	query := `INSERT INTO folders (name, parent_id) VALUES (?, ?)`
	result, err := db.Exec(query, folder.Name, folder.ParentID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	folder.ID = int(id)
	return db.GetFolder(folder.ID, folder)
}

func (db *DB) GetFolder(id int, folder *models.Folder) error {
	query := `SELECT id, name, parent_id, created_at, updated_at FROM folders WHERE id = ?`
	return db.QueryRow(query, id).Scan(&folder.ID, &folder.Name, &folder.ParentID, &folder.CreatedAt, &folder.UpdatedAt)
}

func (db *DB) GetFolders() ([]models.Folder, error) {
	query := `SELECT id, name, parent_id, created_at, updated_at FROM folders ORDER BY name`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.Folder
	for rows.Next() {
		var folder models.Folder
		if err := rows.Scan(&folder.ID, &folder.Name, &folder.ParentID, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, rows.Err()
}

func (db *DB) UpdateFolder(folder *models.Folder) error {
	query := `UPDATE folders SET name = ?, parent_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := db.Exec(query, folder.Name, folder.ParentID, folder.ID)
	return err
}

func (db *DB) DeleteFolder(id int) error {
	query := `DELETE FROM folders WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) CreateRequest(request *models.Request) error {
	headers, err := json.Marshal(request.Headers)
	if err != nil {
		return err
	}

	query := `INSERT INTO requests (name, folder_id, method, url, headers, body) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, request.Name, request.FolderID, request.Method, request.URL, string(headers), request.Body)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	request.ID = int(id)
	return db.GetRequest(request.ID, request)
}

func (db *DB) GetRequest(id int, request *models.Request) error {
	query := `SELECT id, name, folder_id, method, url, headers, body, created_at, updated_at FROM requests WHERE id = ?`
	var headersStr string
	err := db.QueryRow(query, id).Scan(&request.ID, &request.Name, &request.FolderID, &request.Method, &request.URL, &headersStr, &request.Body, &request.CreatedAt, &request.UpdatedAt)
	if err != nil {
		return err
	}

	if headersStr != "" {
		return json.Unmarshal([]byte(headersStr), &request.Headers)
	}
	request.Headers = make(map[string]string)
	return nil
}

func (db *DB) GetRequests() ([]models.Request, error) {
	query := `SELECT id, name, folder_id, method, url, headers, body, created_at, updated_at FROM requests ORDER BY updated_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request
	for rows.Next() {
		var request models.Request
		var headersStr string
		if err := rows.Scan(&request.ID, &request.Name, &request.FolderID, &request.Method, &request.URL, &headersStr, &request.Body, &request.CreatedAt, &request.UpdatedAt); err != nil {
			return nil, err
		}

		if headersStr != "" {
			if err := json.Unmarshal([]byte(headersStr), &request.Headers); err != nil {
				return nil, err
			}
		} else {
			request.Headers = make(map[string]string)
		}

		requests = append(requests, request)
	}

	return requests, rows.Err()
}

func (db *DB) UpdateRequest(request *models.Request) error {
	headers, err := json.Marshal(request.Headers)
	if err != nil {
		return err
	}

	query := `UPDATE requests SET name = ?, folder_id = ?, method = ?, url = ?, headers = ?, body = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = db.Exec(query, request.Name, request.FolderID, request.Method, request.URL, string(headers), request.Body, request.ID)
	return err
}

func (db *DB) DeleteRequest(id int) error {
	query := `DELETE FROM requests WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
