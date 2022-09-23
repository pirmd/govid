package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Note represents a file that rvid knows how to interact with.
type Note struct {
	Filename string
	Content  string
}

// WebApp represents rvid application.
type WebApp struct {
	Storage   WriteFS
	Templates *template.Template
}

// NewWebApp creates a new WebApp providing rvid services. It intaracts with
// files found in noteFS and uses html templates from 'tmpl' folder found
// within tmplFs.
func NewWebApp(noteFs WriteFS, tmplFs fs.FS) *WebApp {
	return &WebApp{
		Storage: noteFs,
		Templates: template.Must(
			template.New("rvid").ParseFS(tmplFs, "tmpl/edit.html.gotmpl"),
		),
	}
}

func (app *WebApp) openNote(filename string) (*Note, error) {
	fi, err := fs.Stat(app.Storage, filename)
	if err != nil {
		if perr, ok := err.(*fs.PathError); ok {
			if errors.Is(perr.Err, fs.ErrNotExist) {
				return &Note{
					Filename: filename,
				}, nil
			}
		}

		return nil, err
	}

	if fi.IsDir() {
		return nil, errors.New("non supported type")
	}

	content, err := fs.ReadFile(app.Storage, filename)
	if err != nil {
		return nil, err
	}

	if mimetype := http.DetectContentType(content); !strings.HasPrefix(mimetype, "text/") {
		return nil, fmt.Errorf("%s is non supported", mimetype)
	}

	return &Note{
		Filename: filename,
		Content:  string(content),
	}, nil
}

// EditHandler is the http.Handler responsible of note editing.
func (app *WebApp) EditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	note, err := app.openNote(vars["filename"])
	if err != nil {
		log.Printf("Opening note '%s' failed: %v", vars["filename"], err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers to avoid caching provided data
	// https://stackoverflow.com/questions/49547/how-do-we-control-web-page-caching-across-all-browsers/2068407#2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")

	if err := app.Templates.ExecuteTemplate(w, "edit.html.gotmpl", note); err != nil {
		log.Printf("Rendering of '%s' failed: %v", vars["filename"], err)
	}
}

// SaveHandler is the http.Handler responsible for saving data to notes.
func (app *WebApp) SaveHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	content := []byte(r.FormValue("content"))

	if mimetype := http.DetectContentType(content); !strings.HasPrefix(mimetype, "text/") {
		log.Printf("Request saving content of mime-type '%s': not supported", mimetype)
		http.Error(w, fmt.Sprintf("%s is non supported", mimetype), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	if err := app.Storage.WriteFile(vars["filename"], content, 0660); err != nil {
		log.Printf("Saving of '%s' failed: %v", vars["filename"], err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
