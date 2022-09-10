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
)

func main() {
	addr := flag.String("address", "localhost:8080", "TCP network address to listen to")
	dir := flag.String("dir", "./notes", "folder that contains notes")
	flag.Parse()

	log.Println("Serving notes from: ", *dir)
	app := NewWebApp(NewDirFS(*dir), tmplFs)

	r := mux.NewRouter()
	r.PathPrefix("/js").Handler(http.FileServer(http.FS(jsFs)))

	r.HandleFunc("/save/{filename}", app.SaveHandler)
	r.HandleFunc("/{filename}", app.EditHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         *addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Println("Starting server on: ", *addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
