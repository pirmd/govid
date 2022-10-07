package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"
)

// Note represents a file that govid knows how to interact with.
type Note struct {
	Filename string
	Content  []byte
}

// Text returns the Note's content as a string, mainly to facilitate inclusion
// into templates.
func (n Note) Text() string {
	return string(n.Content)
}

// WebApp represents govid application.
type WebApp struct {
	Storage   WriteFS
	Templates *template.Template
}

// NewWebApp creates a new WebApp providing govid services. It interacts with
// files found in noteFS and uses html templates from 'tmpl' folder found
// within tmplFs.
func NewWebApp(noteFs WriteFS, tmplFs fs.FS) *WebApp {
	return &WebApp{
		Storage: noteFs,
		Templates: template.Must(
			template.New("govid").ParseFS(tmplFs, "tmpl/edit.html.gotmpl"),
		),
	}
}

// EditHandlerFunc is the http.HandlerFunc responsible of note editing.
func (app *WebApp) EditHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filename := app.sanitizeFilename(r.URL.Path)

	note, err := app.openNote(filename)
	if err != nil {
		log.Printf("opening note '%s' failed: %v", filename, err)
		http.Error(w, fmt.Sprintf("editing '%s' failed", filename), http.StatusInternalServerError)
		return
	}

	if !app.isValidContentType([]byte(note.Content)) {
		log.Printf("editing '%s' denied, not allowed mime-type", filename)
		http.Error(w, fmt.Sprintf("editing '%s' not allowed", filename), http.StatusBadRequest)
		return
	}

	// Set headers to avoid caching provided data
	// https://stackoverflow.com/questions/49547/how-do-we-control-web-page-caching-across-all-browsers/2068407#2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")

	if err := app.Templates.ExecuteTemplate(w, "edit.html.gotmpl", note); err != nil {
		log.Printf("rendering edit template for '%s' failed: %v", filename, err)
	}
}

// SaveHandlerFunc is the http.HandlerFunc responsible for saving data to notes.
func (app *WebApp) SaveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	content := []byte(r.FormValue("content"))
	filename := app.sanitizeFilename(r.URL.Path)

	if !app.isValidContentType(content) {
		log.Printf("saving '%s' denied: not allowed mime-type", filename)
		http.Error(w, fmt.Sprintf("saving '%s' not allowed", filename), http.StatusBadRequest)
		return
	}

	if err := app.Storage.WriteFile(filename, content, 0660); err != nil {
		log.Printf("saving '%s' failed: %v", filename, err)
		http.Error(w, fmt.Sprintf("saving '%s' failed", filename), http.StatusInternalServerError)
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

func (app *WebApp) sanitizeFilename(filename string) string {
	// TODO: check if it is needed, as DirFS might already prevent path
	// traversal risks
	cleaned := path.Clean(path.Join("/", filename))

	// remove the root slash that was added in the previous step and that DirFS is not found of
	return cleaned[1:]
}
