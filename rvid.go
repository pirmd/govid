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
	Content  []byte
}

// Text returns the Note's content as a string, mainly to facilitate inclusion
// into templates.
func (n Note) Text() string {
	return string(n.Content)
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

// EditHandler is the http.Handler responsible of note editing.
func (app *WebApp) EditHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	note, err := app.openNote(vars["filename"])
	if err != nil {
		log.Printf("Opening note '%s' failed: %v", vars["filename"], err)
		http.Error(w, fmt.Sprintf("editing '%s' failed", vars["filename"]), http.StatusInternalServerError)
		return
	}

	if !app.isValidContentType([]byte(note.Content)) {
		log.Printf("editing '%s': access denied, not supported mime-type", vars["filename"])
		http.Error(w, fmt.Sprintf("editing '%s' not supported", vars["filename"]), http.StatusBadRequest)
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

	vars := mux.Vars(r)

	if !app.isValidContentType(content) {
		log.Printf("saving '%s': access denied, not supported mime-type", vars["filename"])
		http.Error(w, fmt.Sprintf("saving '%s' not supported", vars["filename"]), http.StatusBadRequest)
		return
	}

	if err := app.Storage.WriteFile(vars["filename"], content, 0660); err != nil {
		log.Printf("Saving of '%s' failed: %v", vars["filename"], err)
		http.Error(w, fmt.Sprintf("saving '%s' failed", vars["filename"]), http.StatusInternalServerError)
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

	return &Note{
		Filename: filename,
		Content:  content,
	}, nil
}

func (app *WebApp) isValidContentType(content []byte) bool {
	mimetype := http.DetectContentType(content)
	return strings.HasPrefix(mimetype, "text/")
}
