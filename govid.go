package main

import (
	"bytes"
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	// browserTemplate is the name of template in tmplFS to use to browse a
	// folder.
	browserTemplate = "browser.html.gotmpl"

	// editorTemplate is the name of template in tmplFS to use to render note
	// for edition.
	editorTemplate = "editor.html.gotmpl"

	// maxSize defines the limit in size of file to edit.
	maxSize int64 = 1 << 20
)

var (
	//go:embed templates/*.gotmpl
	tmplFS embed.FS
)

// File represents a file or a folder that govid knows how to interact with.
type File struct {
	Filename  string
	Content   string
	Entries   []os.DirEntry
	URLPrefix string
}

// Name returns File base name.
func (f File) Name() string {
	return path.Base(f.Filename)
}

// PathComponents splits File's Filename into its individual path components.
func (f File) PathComponents() []*File {
	p := splitPath(f.Filename)
	c := make([]*File, len(p))

	for i := range p {
		if i == 0 {
			c[i] = &File{
				Filename:  p[i],
				URLPrefix: f.URLPrefix,
			}
			continue
		}

		c[i] = &File{
			Filename:  path.Join(c[i-1].Filename, p[i]),
			URLPrefix: f.URLPrefix,
		}
	}

	return c
}

// URL returns the IRL where File can be accessed from
func (f File) URL() string {
	return path.Join(f.URLPrefix, f.Filename)
}

// WebApp represents govid application.
type WebApp struct {
	URLPrefix string
	RootDir   string
	Templates *template.Template
}

// NewWebApp creates a new WebApp providing govid services for notes found in dir.
// Notes content are rendered using templates from tmplFS's 'templates' subdir.
func NewWebApp(dir string, prefix string) *WebApp {
	return &WebApp{
		RootDir:   dir,
		URLPrefix: prefix,
		Templates: template.Must(
			template.New("govid").ParseFS(tmplFS, "templates/*.gotmpl"),
		),
	}
}

// GetHandlerFunc is the http.HandlerFunc responsible of getting resources.
func (app WebApp) GetHandlerFunc(w http.ResponseWriter, r *http.Request) {
	filepath, _ := strings.CutPrefix(r.URL.Path, app.URLPrefix)

	if !app.isValidPathname(filepath) {
		log.Printf("cannot access '%s': invalid path name", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	fullpath := app.fullpath(filepath)
	fi, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			// Edit a new File (ie: no content)
			file := File{
				Filename:  filepath,
				URLPrefix: app.URLPrefix,
			}
			app.serveTemplate(w, editorTemplate, file)
			return
		}

		log.Printf("cannot access '%s': %v", filepath, err)
		raiseHTTPError(w, err)
		return
	}

	if fi.IsDir() {
		entries, err := os.ReadDir(fullpath)
		if err != nil {
			log.Printf("cannot access folder '%s': %v", filepath, err)
			raiseHTTPError(w, err)
			return
		}

		validEntries := []os.DirEntry{}
		for _, entry := range entries {
			if app.isValidPathname(entry.Name()) {
				validEntries = append(validEntries, entry)
			}
		}

		dir := File{
			Filename:  filepath,
			Entries:   validEntries,
			URLPrefix: app.URLPrefix,
		}
		app.serveTemplate(w, browserTemplate, dir)
		return
	}

	if fi.Size() > maxSize {
		log.Printf("cannot access file '%s': size too big", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	//#nosec G304 -- fullpath is sanitized using app.fullpath
	content, err := os.ReadFile(fullpath)
	if err != nil {
		log.Printf("cannot access file '%s': %v", filepath, err)
		raiseHTTPError(w, err)
		return
	}

	if !app.isValidContentType(content) {
		log.Printf("cannot access file '%s': not allowed mime-type", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	file := File{
		Filename:  filepath,
		Content:   string(content),
		URLPrefix: app.URLPrefix,
	}
	app.serveTemplate(w, editorTemplate, file)
}

// SaveHandlerFunc is the http.HandlerFunc responsible for saving data to notes.
func (app WebApp) SaveHandlerFunc(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	filepath, _ := strings.CutPrefix(r.URL.Path, app.URLPrefix)
	if !app.isValidPathname(filepath) {
		log.Printf("saving to '%s' failed: invalid  path name", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	content := []byte(r.FormValue("content"))
	if !app.isValidContentType(content) {
		log.Printf("saving to '%s' failed: not allowed mime-type", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	// replace CR LF \r\n (windows) with LF \n (unix)
	content = bytes.Replace(content, []byte{13, 10}, []byte{10}, -1)

	fullpath := app.fullpath(filepath)

	fi, err := os.Stat(fullpath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		raiseHTTPError(w, err)
		return
	}

	if fi != nil && fi.IsDir() {
		log.Printf("saving to '%s' failed: is a directory", filepath)
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}

	//#nosec G301 -- O777 is here to let user's umask do its job
	if err := os.MkdirAll(path.Dir(fullpath), 0777); err != nil {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		raiseHTTPError(w, err)
		return
	}

	//#nosec G306 -- O666 is here to let user's umask do its job
	if err := os.WriteFile(fullpath, content, 0666); err != nil {
		log.Printf("saving to '%s' failed: %v", filepath, err)
		raiseHTTPError(w, err)
	}
}

func (app WebApp) serveTemplate(w http.ResponseWriter, tmplName string, file File) {
	// Set headers to avoid caching provided data
	// https://stackoverflow.com/questions/49547/how-do-we-control-web-page-caching-across-all-browsers/2068407#2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	w.Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
	w.Header().Set("Expires", "0")                                         // Proxies.

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")

	buf := new(bytes.Buffer)
	if err := app.Templates.ExecuteTemplate(buf, tmplName, file); err != nil {
		log.Printf("rendering edit template for '%s' failed: %v", file.Filename, err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("rendering edit template for '%s' failed: %v", file.Filename, err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func (app WebApp) isValidPathname(filepath string) bool {
	if containsDotDot(filepath) {
		return false
	}

	if containsHiddenFile(filepath) {
		return false
	}

	return true
}

func (app WebApp) isValidContentType(content []byte) bool {
	mimetype := http.DetectContentType(content)
	return strings.HasPrefix(mimetype, "text/")
}

func (app WebApp) fullpath(filepath string) string {
	return path.Join(app.RootDir, path.Clean("/"+filepath))
}

func raiseHTTPError(w http.ResponseWriter, err error) {
	switch {
	case os.IsNotExist(err):
		http.Error(w, "404 Page not found", http.StatusNotFound)
	case os.IsPermission(err):
		http.Error(w, "403 Forbidden", http.StatusForbidden)
	default:
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func splitPath(filepath string) []string {
	return strings.FieldsFunc(filepath, func(c rune) bool {
		return c == '/'
	})
}

func containsDotDot(filepath string) bool {
	if !strings.Contains(filepath, "..") {
		return false
	}
	for _, p := range splitPath(filepath) {
		if p == ".." {
			return true
		}
	}
	return false
}

func containsHiddenFile(filepath string) bool {
	if !strings.Contains(filepath, ".") {
		return false
	}
	for _, p := range splitPath(filepath) {
		if p == "." || p == ".." {
			continue
		}

		if strings.HasPrefix(p, ".") {
			return true
		}
	}
	return false
}
