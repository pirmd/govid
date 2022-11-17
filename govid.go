package main

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"
)

const (
	// editTemplateName is the name of template in tmplFS to use to render note
	// for edition.
	editTemplateName = "edit.html.gotmpl"

	// maxNoteSize defines the acceptable limit in size for note
	maxNoteSize int64 = 1 << 20
)

var (
	//go:embed templates/*.gotmpl
	tmplFS embed.FS
)

// Note represents a file that govid knows how to interact with.
type Note struct {
	Filepath string
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

// NewWebApp creates a new WebApp providing govid services for notes found in notesdir.
// Notes content are rendered using templates from tmplFS's 'templates' subdir.
func NewWebApp(notesdir string) *WebApp {
	return &WebApp{
		NotesDir: notesdir,
		Templates: template.Must(
			template.New("govid").ParseFS(tmplFS, "templates/*.gotmpl"),
		),
	}
}

// EditHandlerFunc is the http.HandlerFunc responsible of note editing.
func (app WebApp) EditHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filepath := r.URL.Path

	if !app.isValidPathname(filepath) {
		log.Printf("editing '%s' failed: invalid path name", filepath)
		http.Error(w, "edit not possible", http.StatusBadRequest)
		return
	}

	note, err := app.openNote(filepath)
	if err != nil {
		log.Printf("editing '%s' failed: %v", filepath, err)
		http.Error(w, "edit not possible", http.StatusBadRequest)
		return
	}

	if !app.isValidContentType([]byte(note.Content)) {
		log.Printf("editing '%s' failed: not allowed mime-type", filepath)
		http.Error(w, "edit not supported", http.StatusBadRequest)
		return
	}

	// Set headers to avoid caching provided data
	// https://stackoverflow.com/questions/49547/how-do-we-control-web-page-caching-across-all-browsers/2068407#2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")

	buf := new(bytes.Buffer)
	if err := app.Templates.ExecuteTemplate(buf, editTemplateName, note); err != nil {
		log.Printf("rendering edit template for '%s' failed: %v", filepath, err)
		http.Error(w, "rendering note content failed", http.StatusInternalServerError)
	}

	if _, err = buf.WriteTo(w); err != nil {
		log.Printf("rendering edit template for '%s' failed: %v", filepath, err)
		http.Error(w, "rendering note content failed", http.StatusInternalServerError)
	}
}

// SaveHandlerFunc is the http.HandlerFunc responsible for saving data to notes.
func (app WebApp) SaveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxNoteSize)

	filepath := r.URL.Path
	if !app.isValidPathname(filepath) {
		log.Printf("saving to '%s' failed: invalid  path name", filepath)
		http.Error(w, "save not possible", http.StatusBadRequest)
		return
	}

	content := []byte(r.FormValue("content"))
	if !app.isValidContentType(content) {
		log.Printf("saving to '%s' failed: not allowed mime-type", filepath)
		http.Error(w, "save not supported", http.StatusBadRequest)
		return
	}

	fullpath := app.fullpath(filepath)

	fi, err := os.Stat(fullpath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		http.Error(w, "save not supported", http.StatusInternalServerError)
		return
	}

	if fi != nil && fi.IsDir() {
		log.Printf("saving to '%s' failed: is a directory", filepath)
		http.Error(w, "save not supported", http.StatusBadRequest)
		return
	}

	//#nosec G301 -- O777 is here to let user's umask do its job
	if err := os.MkdirAll(path.Dir(fullpath), 0777); err != nil {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		http.Error(w, "save failed", http.StatusInternalServerError)
		return
	}

	//#nosec G306 -- O666 is here to let user's umask do its job
	if err := os.WriteFile(fullpath, content, 0666); err != nil {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		http.Error(w, "save failed", http.StatusInternalServerError)
	}
}

func (app WebApp) openNote(filepath string) (*Note, error) {
	fullpath := app.fullpath(filepath)

	fi, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Note{
				Filepath: filepath,
			}, nil
		}

		return nil, err
	}

	if fi.IsDir() {
		return nil, errors.New("is a directory")
	}

	if fi.Size() > maxNoteSize {
		return nil, errors.New("is too big")
	}

	content, err := os.ReadFile(fullpath) //#nosec G304 -- fullpath is sanitized using app.fullpath
	if err != nil {
		return nil, err
	}

	return &Note{
		Filepath: filepath,
		Content:  content,
	}, nil
}

func (app WebApp) isValidPathname(filepath string) bool {
	if path.Clean("/"+filepath) != filepath {
		return false
	}

	for _, c := range filepath {
		if !unicode.IsPrint(c) {
			return false
		}
	}

	return true
}

func (app WebApp) isValidContentType(content []byte) bool {
	mimetype := http.DetectContentType(content)
	return strings.HasPrefix(mimetype, "text/")
}

func (app WebApp) fullpath(filepath string) string {
	return path.Join(app.NotesDir, path.Clean("/"+filepath))
}
