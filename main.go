package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	//go:embed tmpl
	tmplFs embed.FS

	//go:embed js
	jsFs embed.FS

	//go:embed css
	cssFs embed.FS
)

func main() {
	addr := flag.String("address", "localhost:8080", "TCP network address to listen to")
	dir := flag.String("dir", "./notes", "folder that contains notes")
	htpasswdfile := flag.String("htpasswd", "", "path to htpasswd-like file containing access credentials expected to use bcrypt-based password hash. (default no authentication)")
	flag.Parse()

	authnHandler := noopHandler
	if *htpasswdfile != "" {
		htpasswd, err := NewHtpasswdFromFile(*htpasswdfile)
		if err != nil {
			log.Fatalf("Fail to parse htpasswd credentials: %v", err)
		}

		log.Printf("Authenticate using credentials from %s [%d user(s)]", *htpasswdfile, len(htpasswd))
		authnHandler = htpasswd.BasicAuthHandler
	}

	log.Println("Serving notes from: ", *dir)
	app := NewWebApp(NewDirFS(*dir), tmplFs)

	r := mux.NewRouter()
	r.PathPrefix("/js").Handler(http.FileServer(http.FS(jsFs)))
	r.PathPrefix("/css").Handler(http.FileServer(http.FS(cssFs)))

	r.HandleFunc("/save/{filename}", authnHandler(app.SaveHandler))
	r.HandleFunc("/{filename}", authnHandler(app.EditHandler))

	srv := &http.Server{
		Handler:           r,
		Addr:              *addr,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Println("Starting server on: ", *addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
