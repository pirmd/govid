package main

import (
	"io/fs"
	"os"
	"runtime"
)

// WriteFS extends fs.FS to provide basic write operation.
type WriteFS interface {
	fs.FS

	WriteFile(name string, data []byte, perm os.FileMode) error
}

// DirFS implements WriteFS for file operations within a file-system's folder.
type DirFS struct {
	root string

	fs.FS
}

// NewDirFS creates a DirFS rooted at root. NewDirFS does not create root
// if it does not exist nor verify that it is readable/writeable.
func NewDirFS(dir string) *DirFS {
	return &DirFS{
		root: dir,
		FS:   os.DirFS(dir),
	}
}

// WriteFile writes data to the named file, creating it if necessary.
func (dir *DirFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	path, err := dir.realPath(name)
	if err != nil {
		return &os.PathError{Op: "write", Path: name, Err: err}
	}

	return os.WriteFile(path, data, perm)
}

// realPath returns the "real" path of a file within a dir.
func (dir *DirFS) realPath(name string) (string, error) {
	if !fs.ValidPath(name) || runtime.GOOS == "windows" && containsAny(name, `\:`) {
		return "", os.ErrInvalid
	}

	return dir.root + "/" + name, nil
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
