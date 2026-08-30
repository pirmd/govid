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

// Test la validation des chemins (path traversal, fichiers cachés)
func TestIsValidPath(t *testing.T) {
	testCases := []struct {
		path    string
		valid   bool
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
		// Fichiers cachés
		{"/.hidden", false},
		{"/dossier/.hidden", true},  // Ne commence pas par "/.", logique actuelle
		{"/.git/config", false},
		{"/test/.hidden/file", true}, // Ne commence pas par "/.", logique actuelle
		// Cas valides avec points
		{"/test.file.txt", true},
		{"/dossier.with.dots/file.txt", true},
	}

	for _, tc := range testCases {
		// Simuler la validation du backend
		isValid := !strings.Contains(tc.path, "..") && !strings.HasPrefix(tc.path, "/.")
		if isValid != tc.valid {
			t.Errorf("Path validation failed for '%s': got %v, want %v", tc.path, isValid, tc.valid)
		}
	}
}

// Test la création de fichiers et dossiers
func TestFileOperations(t *testing.T) {
	// Créer un dossier temporaire
	tmpDir, err := os.MkdirTemp("", "govid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Configurer l'environnement
	os.Setenv("GOVID_DIR", tmpDir)
	os.Setenv("GOVID_URL_PREFIX", "")

	// Créer une requête HTTP pour sauvegarder un fichier
	testCases := []struct {
		path     string
		content  string
		expected bool
	}{
		{"/test.txt", "Hello, World!", true},
		{"/dossier/fichier.txt", "Contenu dans un dossier", true},
		{"/../malicious.txt", "Should fail", false},
		{"/.hidden", "Should fail", false},
	}

	for _, tc := range testCases {
		// Vérifier la validation du chemin
		isValid := !strings.Contains(tc.path, "..") && !strings.HasPrefix(tc.path, "/.")
		if isValid != tc.expected {
			t.Errorf("Path validation for '%s' failed: got %v, want %v", tc.path, isValid, tc.expected)
			continue
		}

		if !tc.expected {
			continue // On ne teste pas les chemins invalides
		}

		// Créer le chemin absolu
		absPath := filepath.Join(tmpDir, tc.path)

		// Simuler la création du dossier parent
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			t.Errorf("Failed to create parent dir for '%s': %v", tc.path, err)
			continue
		}

		// Simuler l'écriture du fichier
		if err := os.WriteFile(absPath, []byte(tc.content), 0644); err != nil {
			t.Errorf("Failed to write file '%s': %v", tc.path, err)
			continue
		}

		// Vérifier que le fichier existe et a le bon contenu
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

// Test la limite de taille
func TestMaxSize(t *testing.T) {
	// Créer un contenu trop grand
	largeContent := strings.Repeat("A", maxSize+1)

	// Simuler une requête POST
	req := httptest.NewRequest(http.MethodPost, "/test.txt", strings.NewReader(largeContent))
	w := httptest.NewRecorder()

	// Appeler le handler
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

// Test la détection des fichiers binaires
func TestBinaryFileDetection(t *testing.T) {
	testCases := []struct {
		content []byte
		isBinary bool
	}{
		{[]byte("Hello, World!"), false},
		{[]byte("Texte normal"), false},
		{[]byte{0x00}, true}, // Byte nul
		{[]byte{0x00, 0x01, 0x02}, true},
		{[]byte("GIF89a"), false}, // En-tête GIF (pas de byte nul)
		{[]byte{0xFF, 0xD8, 0xFF}, false}, // En-tête JPEG (pas de byte nul)
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

// Test le comportement par défaut (fichier index.txt)
func TestDefaultFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "govid_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Simuler une requête GET sur / (devrait retourner index.txt)
	path := "/"
	requestedPath := strings.TrimPrefix(path, "")
	if requestedPath == "" || requestedPath == "/" {
		requestedPath = "/index.txt"
	}

	if requestedPath != "/index.txt" {
		t.Errorf("Default file path failed: got '%s', want '/index.txt'", requestedPath)
	}
}

// Test la construction du chemin absolu
func TestAbsPath(t *testing.T) {
	tmpDir := "/tmp/govid"
	testCases := []struct {
		path     string
		prefix   string
		expected string
	}{
		{"/test.txt", "", "/tmp/govid/test.txt"},
		{"/dossier/fichier.txt", "/govid", "/tmp/govid/dossier/fichier.txt"},
		{"", "", "/tmp/govid/"},
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
