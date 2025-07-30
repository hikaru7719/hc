package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hc/hc/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

// intPtr returns a pointer to an int
func intPtr(i int) *int {
	return &i
}

// setupTestDB creates a test database
func setupTestDB(t *testing.T) *DB {
	t.Helper()

	// Create temp directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Override getDBPath for testing
	oldGetDBPath := getDBPath
	getDBPath = func() (string, error) {
		return dbPath, nil
	}
	t.Cleanup(func() {
		getDBPath = oldGetDBPath
	})

	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})

	return db
}

func TestInitDB(t *testing.T) {
	db := setupTestDB(t)

	// Test that tables exist
	tables := []string{"folders", "requests"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}

	// Test ping
	if err := db.Ping(); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestWithTx(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Test successful transaction
	var result int
	err := db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO folders (name) VALUES (?)", "Test Folder")
		if err != nil {
			return err
		}

		err = tx.QueryRow("SELECT COUNT(*) FROM folders").Scan(&result)
		return err
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected 1 folder, got %d", result)
	}

	// Test rollback on error
	err = db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO folders (name) VALUES (?)", "Rollback Folder")
		if err != nil {
			return err
		}

		// Force an error
		return sql.ErrTxDone
	})

	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback
	var count int
	db.QueryRow("SELECT COUNT(*) FROM folders WHERE name = ?", "Rollback Folder").Scan(&count)
	if count != 0 {
		t.Errorf("Expected rollback, but found %d folders", count)
	}
}

// Folder tests

func TestCreateFolder(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name    string
		folder  *models.Folder
		wantErr bool
	}{
		{
			name: "Create simple folder",
			folder: &models.Folder{
				Name: "Test Folder",
			},
			wantErr: false,
		},
		{
			name: "Create folder with parent",
			folder: &models.Folder{
				Name:     "Child Folder",
				ParentID: intPtr(1),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.CreateFolder(tt.folder)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.folder.ID == 0 {
					t.Error("Expected folder ID to be set")
				}

				if tt.folder.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}

				if tt.folder.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestGetFolder(t *testing.T) {
	db := setupTestDB(t)

	// Create a folder first
	originalFolder := &models.Folder{
		Name: "Test Folder",
	}
	if err := db.CreateFolder(originalFolder); err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Get existing folder",
			id:      originalFolder.ID,
			wantErr: false,
		},
		{
			name:    "Get non-existing folder",
			id:      9999,
			wantErr: true,
			errMsg:  "folder not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var folder models.Folder
			err := db.GetFolder(tt.id, &folder)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetFolder() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				if folder.Name != originalFolder.Name {
					t.Errorf("Got folder name %v, want %v", folder.Name, originalFolder.Name)
				}
			}
		})
	}
}

func TestGetFolders(t *testing.T) {
	db := setupTestDB(t)

	// Create test folders
	folderNames := []string{"Alpha", "Beta", "Gamma"}
	for _, name := range folderNames {
		folder := &models.Folder{Name: name}
		if err := db.CreateFolder(folder); err != nil {
			t.Fatalf("Failed to create folder %s: %v", name, err)
		}
	}

	folders, err := db.GetFolders()
	if err != nil {
		t.Fatalf("GetFolders() error = %v", err)
	}

	if len(folders) != len(folderNames) {
		t.Errorf("Got %d folders, want %d", len(folders), len(folderNames))
	}

	// Check ordering (should be alphabetical)
	for i, folder := range folders {
		if folder.Name != folderNames[i] {
			t.Errorf("Folder %d: got name %v, want %v", i, folder.Name, folderNames[i])
		}
	}
}

func TestUpdateFolder(t *testing.T) {
	db := setupTestDB(t)

	// Create initial folder
	createdFolder := &models.Folder{Name: "Original Name"}
	if err := db.CreateFolder(createdFolder); err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	// SQLite datetime has second precision, so wait at least 1 second
	time.Sleep(1 * time.Second)

	// Update the folder
	updateFolder := &models.Folder{
		ID:   createdFolder.ID,
		Name: "Updated Name",
	}
	err := db.UpdateFolder(updateFolder)
	if err != nil {
		t.Errorf("UpdateFolder() error = %v", err)
	}

	// Get updated folder
	var updatedFolder models.Folder
	if err := db.GetFolder(createdFolder.ID, &updatedFolder); err != nil {
		t.Fatalf("Failed to get updated folder: %v", err)
	}

	if updatedFolder.Name != "Updated Name" {
		t.Errorf("Folder name not updated: got %v, want %v", updatedFolder.Name, "Updated Name")
	}

	if !updatedFolder.CreatedAt.Equal(createdFolder.CreatedAt) {
		t.Error("CreatedAt should not change on update")
	}

	// SQLite timestamps have second precision
	timeDiff := updatedFolder.UpdatedAt.Sub(createdFolder.UpdatedAt)
	if timeDiff < 1*time.Second {
		t.Errorf("UpdatedAt should be at least 1 second after original. Diff: %v, Original: %v, Updated: %v",
			timeDiff, createdFolder.UpdatedAt, updatedFolder.UpdatedAt)
	}

	// Test updating non-existent folder
	nonExistent := &models.Folder{ID: 9999, Name: "Ghost"}
	err = db.UpdateFolder(nonExistent)
	if err == nil || err.Error() != "folder not found" {
		t.Errorf("Expected 'folder not found' error, got %v", err)
	}
}

func TestDeleteFolder(t *testing.T) {
	db := setupTestDB(t)

	// Create folder
	folder := &models.Folder{Name: "To Delete"}
	if err := db.CreateFolder(folder); err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	// Delete folder
	err := db.DeleteFolder(folder.ID)
	if err != nil {
		t.Errorf("DeleteFolder() error = %v", err)
	}

	// Verify deletion
	var deleted models.Folder
	err = db.GetFolder(folder.ID, &deleted)
	if err == nil || err.Error() != "folder not found" {
		t.Errorf("Expected folder to be deleted, got %v", err)
	}

	// Test deleting non-existent folder
	err = db.DeleteFolder(9999)
	if err == nil || err.Error() != "folder not found" {
		t.Errorf("Expected 'folder not found' error, got %v", err)
	}
}

// Request tests

func TestCreateRequest(t *testing.T) {
	db := setupTestDB(t)

	// Create a folder for the request
	folder := &models.Folder{Name: "Request Folder"}
	if err := db.CreateFolder(folder); err != nil {
		t.Fatalf("Failed to create folder: %v", err)
	}

	tests := []struct {
		name    string
		request *models.Request
		wantErr bool
	}{
		{
			name: "Create simple request",
			request: &models.Request{
				Name:     "Test Request",
				Method:   "GET",
				URL:      "https://example.com",
				Headers:  map[string]string{"Content-Type": "application/json"},
				Body:     `{"test": true}`,
				FolderID: &folder.ID,
			},
			wantErr: false,
		},
		{
			name: "Create request without folder",
			request: &models.Request{
				Name:    "No Folder Request",
				Method:  "POST",
				URL:     "https://example.com/api",
				Headers: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.CreateRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.request.ID == 0 {
					t.Error("Expected request ID to be set")
				}

				if tt.request.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}

				if tt.request.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestGetRequest(t *testing.T) {
	db := setupTestDB(t)

	// Create a request
	originalRequest := &models.Request{
		Name:    "Test Request",
		Method:  "GET",
		URL:     "https://example.com",
		Headers: map[string]string{"X-Test": "true"},
		Body:    "test body",
	}
	if err := db.CreateRequest(originalRequest); err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	tests := []struct {
		name    string
		id      int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Get existing request",
			id:      originalRequest.ID,
			wantErr: false,
		},
		{
			name:    "Get non-existing request",
			id:      9999,
			wantErr: true,
			errMsg:  "request not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request models.Request
			err := db.GetRequest(tt.id, &request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("GetRequest() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				if request.Name != originalRequest.Name {
					t.Errorf("Got request name %v, want %v", request.Name, originalRequest.Name)
				}

				if request.Headers["X-Test"] != "true" {
					t.Errorf("Headers not properly deserialized")
				}
			}
		})
	}
}

func TestGetRequests(t *testing.T) {
	db := setupTestDB(t)

	// Create test requests with different timestamps
	for i := 0; i < 3; i++ {
		request := &models.Request{
			Name:   fmt.Sprintf("Request %d", i),
			Method: "GET",
			URL:    fmt.Sprintf("https://example.com/%d", i),
		}
		if err := db.CreateRequest(request); err != nil {
			t.Fatalf("Failed to create request %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	requests, err := db.GetRequests()
	if err != nil {
		t.Fatalf("GetRequests() error = %v", err)
	}

	if len(requests) != 3 {
		t.Errorf("Got %d requests, want 3", len(requests))
	}

	// Check ordering (should be by updated_at DESC)
	for i := 0; i < len(requests)-1; i++ {
		if requests[i].UpdatedAt.Before(requests[i+1].UpdatedAt) {
			t.Error("Requests not ordered by updated_at DESC")
		}
	}
}

func TestUpdateRequest(t *testing.T) {
	db := setupTestDB(t)

	// Create initial request
	request := &models.Request{
		Name:    "Original Request",
		Method:  "GET",
		URL:     "https://example.com",
		Headers: map[string]string{"X-Original": "true"},
	}
	if err := db.CreateRequest(request); err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Update request
	request.Name = "Updated Request"
	request.Method = "POST"
	request.Headers = map[string]string{"X-Updated": "true"}

	err := db.UpdateRequest(request)
	if err != nil {
		t.Errorf("UpdateRequest() error = %v", err)
	}

	// Verify update
	var updated models.Request
	if err := db.GetRequest(request.ID, &updated); err != nil {
		t.Fatalf("Failed to get updated request: %v", err)
	}

	if updated.Name != "Updated Request" {
		t.Errorf("Request name not updated: got %v, want %v", updated.Name, "Updated Request")
	}

	if updated.Method != "POST" {
		t.Errorf("Request method not updated: got %v, want %v", updated.Method, "POST")
	}

	if updated.Headers["X-Updated"] != "true" {
		t.Error("Headers not properly updated")
	}

	// Test updating non-existent request
	nonExistent := &models.Request{ID: 9999, Name: "Ghost", Method: "GET", URL: "https://ghost.com"}
	err = db.UpdateRequest(nonExistent)
	if err == nil || err.Error() != "request not found" {
		t.Errorf("Expected 'request not found' error, got %v", err)
	}
}

func TestDeleteRequest(t *testing.T) {
	db := setupTestDB(t)

	// Create request
	request := &models.Request{
		Name:   "To Delete",
		Method: "DELETE",
		URL:    "https://example.com",
	}
	if err := db.CreateRequest(request); err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Delete request
	err := db.DeleteRequest(request.ID)
	if err != nil {
		t.Errorf("DeleteRequest() error = %v", err)
	}

	// Verify deletion
	var deleted models.Request
	err = db.GetRequest(request.ID, &deleted)
	if err == nil || err.Error() != "request not found" {
		t.Errorf("Expected request to be deleted, got %v", err)
	}

	// Test deleting non-existent request
	err = db.DeleteRequest(9999)
	if err == nil || err.Error() != "request not found" {
		t.Errorf("Expected 'request not found' error, got %v", err)
	}
}

// Helper function tests

func TestSerializeHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name:    "Normal headers",
			headers: map[string]string{"Content-Type": "application/json", "X-Test": "true"},
			want:    `{"Content-Type":"application/json","X-Test":"true"}`,
		},
		{
			name:    "Empty headers",
			headers: map[string]string{},
			want:    `{}`,
		},
		{
			name:    "Nil headers",
			headers: nil,
			want:    `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := serializeHeaders(tt.headers)
			if err != nil {
				t.Errorf("serializeHeaders() error = %v", err)
				return
			}

			// Compare as JSON to ignore key ordering
			var gotMap, wantMap map[string]string
			json.Unmarshal([]byte(got), &gotMap)
			json.Unmarshal([]byte(tt.want), &wantMap)

			if len(gotMap) != len(wantMap) {
				t.Errorf("serializeHeaders() = %v, want %v", got, tt.want)
			}

			for k, v := range wantMap {
				if gotMap[k] != v {
					t.Errorf("serializeHeaders() key %s = %v, want %v", k, gotMap[k], v)
				}
			}
		})
	}
}

func TestDeserializeHeaders(t *testing.T) {
	tests := []struct {
		name       string
		headersStr string
		want       map[string]string
		wantErr    bool
	}{
		{
			name:       "Normal headers",
			headersStr: `{"Content-Type":"application/json","X-Test":"true"}`,
			want:       map[string]string{"Content-Type": "application/json", "X-Test": "true"},
			wantErr:    false,
		},
		{
			name:       "Empty string",
			headersStr: "",
			want:       map[string]string{},
			wantErr:    false,
		},
		{
			name:       "Invalid JSON",
			headersStr: `{invalid json}`,
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deserializeHeaders(tt.headersStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("deserializeHeaders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("deserializeHeaders() got %d headers, want %d", len(got), len(tt.want))
				}

				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("deserializeHeaders() key %s = %v, want %v", k, got[k], v)
					}
				}
			}
		})
	}
}
