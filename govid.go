package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxSize = 1 << 20 // 1 MB

func main() {
	// Configuration
	dir := os.Getenv("GOVID_DIR")
	if dir == "" {
		dir = os.Getenv("DOCUMENT_ROOT")
	}
	if dir == "" {
		dir = "./notes" // Default directory (relative to binary)
	}

	prefix := os.Getenv("GOVID_URL_PREFIX")
	if prefix == "" {
		prefix = os.Getenv("SCRIPT_NAME")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Extract file path
		requestedPath := strings.TrimPrefix(r.URL.Path, prefix)
		if requestedPath == "" || requestedPath == "/" {
			requestedPath = "/index.txt" // Default file
		}

		// Security validation: block malicious paths
		if strings.Contains(requestedPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Block hidden files/directories anywhere in path
		pathParts := strings.Split(requestedPath, "/")
		for _, part := range pathParts {
			if strings.HasPrefix(part, ".") && part != "" {
				http.Error(w, "Invalid path", http.StatusBadRequest)
				return
			}
		}

		// Absolute path
		absPath := filepath.Join(dir, requestedPath)
		if !strings.HasPrefix(absPath, filepath.Clean(dir)) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Handle requests
		switch r.Method {
		case http.MethodGet:
			// Read file (or return empty if not exists)
			content, err := os.ReadFile(absPath)
			if err != nil && !os.IsNotExist(err) {
				http.Error(w, "Cannot read file", http.StatusInternalServerError)
				return
			}
			// Set appropriate Content-Type based on file extension
			contentType := "text/plain; charset=UTF-8"
			if strings.HasSuffix(requestedPath, ".html") {
				contentType = "text/html; charset=UTF-8"
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(content) // If file doesn't exist, content is empty

		case http.MethodPost:
			// Read POST content (limited to maxSize)
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
			if err != nil {
				http.Error(w, "Request too large", http.StatusBadRequest)
				return
			}

			// Block binary files (null bytes)
			for _, b := range body {
				if b == 0 {
					http.Error(w, "Binary files not allowed", http.StatusBadRequest)
					return
				}
			}

			// Create parent directories if needed
			if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
				http.Error(w, "Cannot create directory", http.StatusInternalServerError)
				return
			}

			// Write file
			if err := os.WriteFile(absPath, body, 0644); err != nil {
				http.Error(w, "Cannot save file", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// CGI mode (Apache/Nginx) or standalone (dev)
	if os.Getenv("GATEWAY_INTERFACE") != "" {
		http.ListenAndServe(":0", nil)
	} else {
		log.Println("Server running on http://localhost:8080")
		log.Println("Notes directory:", dir)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
