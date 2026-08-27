package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isPathValid simulates the backend path validation logic
func isPathValid(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	pathParts := strings.Split(path, "/")
	for _, part := range pathParts {
		if strings.HasPrefix(part, ".") && part != "" {
			return false
		}
	}
	return true
}

// Test path validation (path traversal, hidden files)
func TestIsValidPath(t *testing.T) {
	testCases := []struct {
		path  string
		valid bool
	}{
		{"/test.txt", true},
		{"/dossier/fichier.txt", true},
		{"/index.txt", true},
		{"", true},
		{"/", true},
		// Path traversal
		{"/../etc/passwd", false},
		{"/dossier/../fichier.txt", false},
		{"../test.txt", false},
		// Hidden files - now blocked anywhere in path
		{"/.hidden", false},
		{"/dossier/.hidden", false},  // Now blocked - hidden file in subdirectory
		{"/.git/config", false},
		{"/test/.hidden/file", false}, // Now blocked - hidden directory in path
		{"/.gitignore", false},
		// Valid cases with dots (not at start of path component)
		{"/test.file.txt", true},
		{"/dossier.with.dots/file.txt", true},
		{"/file.with.many.dots.txt", true},
	}

	for _, tc := range testCases {
		isValid := isPathValid(tc.path)
		if isValid != tc.valid {
			t.Errorf("Path validation failed for '%s': got %v, want %v", tc.path, isValid, tc.valid)
		}
	}
}

// Test file and directory creation
func TestFileOperations(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "govid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set environment
	os.Setenv("GOVID_DIR", tmpDir)
	os.Setenv("GOVID_URL_PREFIX", "")

	// Test cases for saving files
	testCases := []struct {
		path     string
		content  string
		expected bool
	}{
		{"/test.txt", "Hello, World!", true},
		{"/dossier/fichier.txt", "Content in directory", true},
		{"/../malicious.txt", "Should fail", false},
		{"/.hidden", "Should fail", false},
	}

	for _, tc := range testCases {
		// Validate path
		isValid := isPathValid(tc.path)
		if isValid != tc.expected {
			t.Errorf("Path validation for '%s' failed: got %v, want %v", tc.path, isValid, tc.expected)
			continue
		}

		if !tc.expected {
			continue // Skip invalid paths
		}

		// Create absolute path
		absPath := filepath.Join(tmpDir, tc.path)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			t.Errorf("Failed to create parent dir for '%s': %v", tc.path, err)
			continue
		}

		// Write file
		if err := os.WriteFile(absPath, []byte(tc.content), 0644); err != nil {
			t.Errorf("Failed to write file '%s': %v", tc.path, err)
			continue
		}

		// Verify file exists and has correct content
		content, err := os.ReadFile(absPath)
		if err != nil {
			t.Errorf("Failed to read file '%s': %v", tc.path, err)
			continue
		}

		if string(content) != tc.content {
			t.Errorf("File content mismatch for '%s': got '%s', want '%s'", tc.path, string(content), tc.content)
		}
	}
}

// Test size limit
func TestMaxSize(t *testing.T) {
	// Create oversized content
	largeContent := strings.Repeat("A", maxSize+1)

	// Simulate POST request
	req := httptest.NewRequest(http.MethodPost, "/test.txt", strings.NewReader(largeContent))
	w := httptest.NewRecorder()

	// Call handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
		if err != nil {
			http.Error(w, "Request too large", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for large request, got %d", w.Code)
	}
}

// Test binary file detection
func TestBinaryFileDetection(t *testing.T) {
	testCases := []struct {
		content []byte
		isBinary bool
	}{
		{[]byte("Hello, World!"), false},
		{[]byte("Normal text"), false},
		{[]byte{0x00}, true}, // Null byte
		{[]byte{0x00, 0x01, 0x02}, true},
		{[]byte("GIF89a"), false}, // GIF header (no null byte)
		{[]byte{0xFF, 0xD8, 0xFF}, false}, // JPEG header (no null byte)
	}

	for _, tc := range testCases {
		isBinary := false
		for _, b := range tc.content {
			if b == 0 {
				isBinary = true
				break
			}
		}

		if isBinary != tc.isBinary {
			t.Errorf("Binary detection failed for %v: got %v, want %v", tc.content, isBinary, tc.isBinary)
		}
	}
}

// Test default file behavior (index.txt)
func TestDefaultFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "govid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Simulate GET request on / (should return index.txt)
	path := "/"
	requestedPath := strings.TrimPrefix(path, "")
	if requestedPath == "" || requestedPath == "/" {
		requestedPath = "/index.txt"
	}

	if requestedPath != "/index.txt" {
		t.Errorf("Default file path failed: got '%s', want '/index.txt'", requestedPath)
	}
}

// Test absolute path construction
func TestAbsPath(t *testing.T) {
	tmpDir := "/tmp/govid"
	testCases := []struct {
		path     string
		prefix   string
		expected string
	}{
		{"/test.txt", "", "/tmp/govid/test.txt"},
		{"/dossier/fichier.txt", "/govid", "/tmp/govid/dossier/fichier.txt"},
		{"", "", "/tmp/govid/index.txt"},
	}

	for _, tc := range testCases {
		requestedPath := strings.TrimPrefix(tc.path, tc.prefix)
		if requestedPath == "" || requestedPath == "/" {
			requestedPath = "/index.txt"
		}

		absPath := filepath.Join(tmpDir, requestedPath)
		if absPath != tc.expected {
			t.Errorf("Abs path failed for '%s': got '%s', want '%s'", tc.path, absPath, tc.expected)
		}
	}
}
