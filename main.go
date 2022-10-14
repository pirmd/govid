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
)

var (
	myname    = filepath.Base(os.Args[0])
	myversion = "v?.?.?-?" //should be set using: go build -ldflags "-X main.myversion=X.X.X"

	//go:embed static/js/*.js
	//go:embed static/css/*.css
	staticFS embed.FS
)

func main() {
	addr := flag.String("address", "localhost:8888", "TCP network address to listen to")
	htpasswdfile := flag.String("htpasswd", "", "path to htpasswd-like file containing access credentials expected to use bcrypt-based password hash. (default no authentication)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [option...] NOTES_DIR\n", myname)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("invalid number of argument(s)\nRun %s -help", myname)
	}

	notesdir := flag.Arg(0)

	log.Printf("Running %s version %s", myname, myversion)
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

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	mux.Handle("/", loggingHandler(authnHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app := NewWebApp(notesdir)

		switch r.Method {
		case http.MethodGet:
			app.EditHandlerFunc(w, r)

		case http.MethodPost:
			app.SaveHandlerFunc(w, r)

		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method "+r.Method+"is not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	srv := &http.Server{
		Handler:           mux,
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
