package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type ctxKey int

const (
	ctxUsernameKey ctxKey = iota
)

// Htpasswd represents credentials in the htpasswd format relying on a bcrypt
// hash.
type Htpasswd map[string]string

// NewHtpasswdFromFile loads htpasswd credentials from a file.
func NewHtpasswdFromFile(filename string) (Htpasswd, error) {
	r, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()

	return NewHtpasswd(r)
}

// NewHtpasswd loads a htpasswd credentials from an io.Reader.
func NewHtpasswd(r io.Reader) (Htpasswd, error) {
	htpwd := make(map[string]string)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			return nil, errors.New("invalid htpassd: entry format shall be 'username:hash'")
		}

		if _, exists := htpwd[parts[0]]; exists {
			return nil, errors.New("invalid htpasswd: users entry shall be uniq")
		}

		htpwd[parts[0]] = parts[1]

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return htpwd, nil
}

// Authenticate validates whether the couple (user, password) matches one of Htpasswd
// credential.
func (htpwd Htpasswd) Authenticate(username, password string) bool {
	expectedPasswordHash, exists := htpwd[username]
	if !exists {
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(expectedPasswordHash), []byte(password)); err != nil {
		return false
	}

	return true
}

// BasicAuthHandler provides a middleware to authenticate user against Htpasswd
// credentials using basic authentication mechanism.
func (htpwd Htpasswd) BasicAuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			if htpwd.Authenticate(username, password) {
				*r = *r.WithContext(context.WithValue(r.Context(), ctxUsernameKey, username))

				next.ServeHTTP(w, r)
				return
			}
			log.Printf("access failed: verification of user id '%s' not configured", username)
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="govid restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// GetUsernameFromContext retrieves a username value from a Request Context as
// set by Htpasswd once authentication succeed.
// Returns "-" if no username has been set.
func GetUsernameFromContext(r *http.Request) string {
	if username := r.Context().Value(ctxUsernameKey); username != nil {
		return username.(string)
	}

	return "-"
}
