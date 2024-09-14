package main

import (
	"log"
	"net/http"
	"net/http/cgi" //#nosec G504 -- Use Go versions > 1.17
	"os"
)

func main() {
	dir := os.Getenv("GOVID_DIR")
	if dir == "" {
		dir = os.Getenv("DOCUMENT_ROOT")
	}

	prefix := os.Getenv("GOVID_URL_PREFIX")
	if prefix == "" {
		prefix = os.Getenv("SCRIPT_NAME")
	}

	app := NewWebApp(dir, prefix)

	notesHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodOptions:
			w.Header().Set("Allow", "OPTIONS, GET, POST")

		case http.MethodGet:
			app.GetHandlerFunc(w, r)

		case http.MethodPost:
			app.SaveHandlerFunc(w, r)

		default:
			w.Header().Set("Allow", "OPTIONS, GET, POST")
			http.Error(w, "method "+r.Method+" is not allowed", http.StatusMethodNotAllowed)
		}
	})

	if err := cgi.Serve(notesHandler); err != nil {
		log.Fatal(err)
	}
}
