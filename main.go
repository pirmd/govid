package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

var (
	myname    = filepath.Base(os.Args[0])
	myversion = "v?.?.?-?" //should be set using: go build -ldflags "-X main.myversion=X.X.X"

	//go:embed tmpl
	tmplFs embed.FS

	//go:embed js
	jsFs embed.FS

	//go:embed css
	cssFs embed.FS
)

func main() {
	addr := flag.String("address", "localhost:8080", "TCP network address to listen to")
	htpasswdfile := flag.String("htpasswd", "", "path to htpasswd-like file containing access credentials expected to use bcrypt-based password hash. (default no authentication)")

	log.Printf("Running %s version %s", myname, myversion)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [option...] NOTES_DIR\n", myname)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("invalid number of argument(s)\nRun %s -help", myname)
	}

	notesdir := flag.Arg(0)
	log.Println("Serving notes from: ", notesdir)

	authnHandler := noopHandler
	if *htpasswdfile != "" {
		htpasswd, err := NewHtpasswdFromFile(*htpasswdfile)
		if err != nil {
			log.Fatalf("Fail to parse htpasswd credentials: %v", err)
		}

		log.Printf("Authenticate using credentials from %s [%d user(s)]", *htpasswdfile, len(htpasswd))
		authnHandler = htpasswd.BasicAuthHandler
	}

	app := NewWebApp(NewDirFS(notesdir), tmplFs)

	r := mux.NewRouter()
	r.PathPrefix("/js").Handler(http.FileServer(http.FS(jsFs)))
	r.PathPrefix("/css").Handler(http.FileServer(http.FS(cssFs)))

	r.Handle("/{filename}", loggingHandler(authnHandler(app.EditHandler()))).
		Methods(http.MethodGet)
	r.Handle("/{filename}", loggingHandler(authnHandler(app.SaveHandler()))).
		Methods(http.MethodPost)

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
