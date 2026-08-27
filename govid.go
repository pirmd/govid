package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxSize = 1 << 20 // 1 Mo

func main() {
	// Configuration
	dir := os.Getenv("GOVID_DIR")
	if dir == "" {
		dir = os.Getenv("DOCUMENT_ROOT")
	}
	if dir == "" {
		dir = "./notes" // Dossier par défaut (relatif au binaire)
	}

	prefix := os.Getenv("GOVID_URL_PREFIX")
	if prefix == "" {
		prefix = os.Getenv("SCRIPT_NAME")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Extraire le chemin du fichier
		filepath := strings.TrimPrefix(r.URL.Path, prefix)
		if filepath == "" || filepath == "/" {
			filepath = "/index.txt" // Fichier par défaut
		}

		// Validation sécurité : bloquer les chemins malveillants
		if strings.Contains(filepath, "..") || strings.HasPrefix(filepath, "/.") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Chemin absolu
		absPath := filepath.Join(dir, filepath)
		if !strings.HasPrefix(absPath, filepath.Clean(dir)) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		// Gestion des requêtes
		switch r.Method {
		case http.MethodGet:
			// Lire le fichier (ou retourner vide si inexistant)
			content, err := os.ReadFile(absPath)
			if err != nil && !os.IsNotExist(err) {
				http.Error(w, "Cannot read file", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			w.Write(content) // Si le fichier n'existe pas, content est vide

		case http.MethodPost:
			// Lire le contenu POST (limité à maxSize)
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
			if err != nil {
				http.Error(w, "Request too large", http.StatusBadRequest)
				return
			}

			// Bloquer les fichiers binaires (bytes nuls)
			for _, b := range body {
				if b == 0 {
					http.Error(w, "Binary files not allowed", http.StatusBadRequest)
					return
				}
			}

			// Créer les dossiers parents si nécessaire
			if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
				http.Error(w, "Cannot create directory", http.StatusInternalServerError)
				return
			}

			// Écrire le fichier
			if err := os.WriteFile(absPath, body, 0644); err != nil {
				http.Error(w, "Cannot save file", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mode CGI (Apache/Nginx) ou standalone (dev)
	if os.Getenv("GATEWAY_INTERFACE") != "" {
		http.ListenAndServe(":0", nil)
	} else {
		log.Println("🚀 Server running on http://localhost:8080")
		log.Println("📝 Notes directory:", dir)
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}
