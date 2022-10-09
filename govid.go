package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
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
	NotesDir  string
	Templates *template.Template
}

// NewWebApp creates a new WebApp providing govid services. It interacts with
// files found in notedir folder and uses html templates from 'tmpl' folder
// found within tmplFs.
func NewWebApp(notesdir string, tmplFs fs.FS) *WebApp {
	return &WebApp{
		NotesDir: notesdir,
		Templates: template.Must(
			template.New("govid").ParseFS(tmplFs, "tmpl/edit.html.gotmpl"),
		),
	}
}

// EditHandlerFunc is the http.HandlerFunc responsible of note editing.
func (app *WebApp) EditHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filename, err := app.sanitizeFilename(r.URL.Path)
	if err != nil {
		log.Printf("filename '%s' is invalid: %v", filename, err)
		http.Error(w, fmt.Sprintf("editing '%s' failed", filename), http.StatusBadRequest)
		return
	}

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

	filename, err := app.sanitizeFilename(r.URL.Path)
	if err != nil {
		log.Printf("filename '%s' is invalid: %v", filename, err)
		http.Error(w, fmt.Sprintf("saving '%s' failed", filename), http.StatusBadRequest)
		return
	}

	if !app.isValidContentType(content) {
		log.Printf("saving '%s' denied: not allowed mime-type", filename)
		http.Error(w, fmt.Sprintf("saving '%s' not allowed", filename), http.StatusBadRequest)
		return
	}

	fullpath := app.fullpath(filename)
	if err := os.MkdirAll(path.Dir(fullpath), os.ModePerm); err != nil {
		log.Printf("saving '%s' failed: %v", filename, err)
		http.Error(w, fmt.Sprintf("saving '%s' failed", filename), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(fullpath, content, os.ModePerm); err != nil {
		log.Printf("saving '%s' failed: %v", filename, err)
		http.Error(w, fmt.Sprintf("saving '%s' failed", filename), http.StatusInternalServerError)
	}
}

func (app *WebApp) openNote(filename string) (*Note, error) {
	fullpath := app.fullpath(filename)

	fi, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Note{
				Filename: filename,
			}, nil
		}

		return nil, err
	}

	if fi.IsDir() {
		return nil, errors.New("non supported type")
	}

	content, err := os.ReadFile(fullpath) //#nosec G304 -- fullpath is sanitized using app.fullpath
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

func (app *WebApp) sanitizeFilename(filename string) (string, error) {
	cleaned := path.Clean(path.Join("/", filename))[1:]

	if !fs.ValidPath(cleaned) || runtime.GOOS == "windows" && containsAny(cleaned, `\:`) {
		return "", os.ErrInvalid
	}

	return cleaned, nil
}

func (app *WebApp) fullpath(filename string) string {
	// TODO: check if it is overkill as sanitizeFilename shall already have
	// ensure that filename is reasonable
	return path.Join(app.NotesDir, path.Clean("/"+filename))
}

func containsAny(s, chars string) bool {
	for i := 0; i < len(s); i++ {
		for j := 0; j < len(chars); j++ {
			if s[i] == chars[j] {
				return true
			}
		}
	}
	return false
}
