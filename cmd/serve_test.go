package cmd

import (
	"io/fs"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// Mock file system for testing
type mockFS struct{}

func (m *mockFS) Open(name string) (fs.File, error) {
	return nil, os.ErrNotExist
}

func TestServeCommand(t *testing.T) {
	// Test that the serve command is properly configured
	if serveCmd == nil {
		t.Fatal("serveCmd should not be nil")
	}

	if serveCmd.Use != "serve" {
		t.Errorf("Expected Use to be 'serve', got %s", serveCmd.Use)
	}

	if serveCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if serveCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if serveCmd.RunE == nil {
		t.Error("RunE function should be defined")
	}
}

func TestAddToRoot(t *testing.T) {
	rootCmd := &cobra.Command{
		Use: "test",
	}

	// Add serve command
	AddToRoot(rootCmd)

	// Check if serve command was added
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "serve" {
			found = true
			break
		}
	}

	if !found {
		t.Error("serve command was not added to root command")
	}
}

func TestServeCmdFlags(t *testing.T) {
	// Reset port to default
	port = 8080

	// Check if flags are properly defined
	if serveCmd.Flag("port") == nil {
		t.Error("port flag not defined")
	}

	if serveCmd.Flag("port").Shorthand != "p" {
		t.Error("port flag shorthand should be 'p'")
	}

	// Test flag parsing
	serveCmd.Flags().Set("port", "9999")
	if port != 9999 {
		t.Errorf("port flag not set correctly, got %d, want 9999", port)
	}
}

func TestGetFrontendFS(t *testing.T) {
	// Save original
	original := GetFrontendFS
	defer func() {
		GetFrontendFS = original
	}()

	// Test with nil GetFrontendFS
	GetFrontendFS = nil
	if GetFrontendFS != nil {
		t.Error("GetFrontendFS should be nil")
	}

	// Test with custom function
	called := false
	GetFrontendFS = func() (fs.FS, error) {
		called = true
		return &mockFS{}, nil
	}

	fs, err := GetFrontendFS()
	if !called {
		t.Error("GetFrontendFS function was not called")
	}
	if err != nil {
		t.Errorf("GetFrontendFS returned error: %v", err)
	}
	if fs == nil {
		t.Error("GetFrontendFS returned nil FS")
	}
}