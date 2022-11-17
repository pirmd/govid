package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

func TestIsValidPathname(t *testing.T) {
	testCases := []struct {
		in   string
		want bool
	}{
		{"/test1", true},
		{"test1", false},
		{"./test1", false},
		{"/../test1", false},
		{"/../test1", false},
		{"/../test/test1", false},
		{"/test/../test1", false},
		{"/test/../../test1", false},
		{"", false},
	}

	testApp := NewWebApp("root")
	for _, tc := range testCases {
		got := testApp.isValidPathname(tc.in)
		if got != tc.want {
			t.Errorf("isValidPathname failed for '%s'\nGot : %v\nWant: %v\n", tc.in, got, tc.want)
		}
	}
}

func TestFullpath(t *testing.T) {
	testCases := []struct {
		in   string
		want string
	}{
		{"", "root"},
		{"test1", path.Join("root", "test1")},
		{"./test1", path.Join("root", "test1")},
		{"/test1", path.Join("root", "test1")},
		{"../test1", path.Join("root", "test1")},
		{"../test/test1", path.Join("root", "test", "test1")},
		{"test/../test1", path.Join("root", "test1")},
		{"test/../../test1", path.Join("root", "test1")},
	}

	testApp := NewWebApp("root")
	for _, tc := range testCases {
		got := testApp.fullpath(tc.in)
		if got != tc.want {
			t.Errorf("Fullpath failed for '%s'\nGot : %s\nWant: %s\n", tc.in, got, tc.want)
		}
	}
}

func TestEditHandler(t *testing.T) {
	testApp, testNotes := setup(t)

	testCases := []struct {
		inFilename string
		outStatus  int
	}{
		{"world_domination", http.StatusOK},
		{"secret", http.StatusOK},
		{"subdir/todo", http.StatusOK},
		{"newnote", http.StatusOK},
		{"subdir/newnote", http.StatusOK},
		{"newsubdir/newnote", http.StatusOK},
		{"../htpasswd", http.StatusBadRequest},
		{"subdir", http.StatusBadRequest},
		{"", http.StatusBadRequest},
		{"1.gif", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		r := httptest.NewRequest(http.MethodGet, "/"+tc.inFilename, nil)
		w := httptest.NewRecorder()
		testApp.EditHandlerFunc(w, r)

		got := w.Result()
		defer func() {
			if err := got.Body.Close(); err != nil {
				t.Fatalf("couldn't close response body for %s: %v", tc.inFilename, err)
			}
		}()

		if got.StatusCode != tc.outStatus {
			t.Fatalf("Status code for %s failed.\nGot: %v\nWant: %v", tc.inFilename, got.StatusCode, tc.outStatus)
		}

		if got.StatusCode == http.StatusOK {
			body, err := io.ReadAll(got.Body)
			if err != nil {
				t.Fatalf("Fail to read response content for %s: %v", tc.inFilename, err)
			}

			want := new(bytes.Buffer)
			n := &Note{
				"/" + tc.inFilename,
				[]byte(testNotes[tc.inFilename]),
			}
			if err := testApp.Templates.ExecuteTemplate(want, editTemplateName, n); err != nil {
				t.Fatalf("rendering edit template for '%s' failed: %v", tc.inFilename, err)
			}

			if string(body) != want.String() {
				t.Errorf("Response body for %s failed.\nGot : %v\nWant: %v", tc.inFilename, string(body), want.String())
			}
		}
	}
}

func TestSaveHandler(t *testing.T) {
	testApp, testNotes := setup(t)

	testCases := []struct {
		inFilename string
		inContent  string
		outStatus  int
	}{
		{"world_domination", "TestMeIfYouCan", http.StatusOK},
		{"1.gif", "TestMeIfYouCan", http.StatusOK},
		{"subdir/todo", "TestMeIfYouCan", http.StatusOK},
		{"newnote", "TestMeIfYouCan", http.StatusOK},
		{"subdir/newnote", "TestMeIfYouCan", http.StatusOK},
		{"newsubdir/newnote", "TestMeIfYouCan", http.StatusOK},
		{"../htpasswd", "TestMeIfYouCan", http.StatusBadRequest},
		{"subdir", "TestMeIfYouCan", http.StatusBadRequest},
		{"", "TestMeIfYouCan", http.StatusBadRequest},
		{"secret", "GIF89a^A^@^A^@^@ÿ^@,^@^@^@^@^A^@^A^@^@^B^@;", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		r := httptest.NewRequest(http.MethodPost, "/"+tc.inFilename, strings.NewReader("content="+url.QueryEscape(tc.inContent)))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		testApp.SaveHandlerFunc(w, r)

		got := w.Result()
		defer func() {
			if err := got.Body.Close(); err != nil {
				t.Fatalf("couldn't close response body for %s: %v", tc.inFilename, err)
			}
		}()

		if got.StatusCode != tc.outStatus {
			t.Fatalf("Status code for %s failed.\nGot: %v\nWant: %v", tc.inFilename, got.StatusCode, tc.outStatus)
		}

		if got.StatusCode == http.StatusOK {
			content, err := os.ReadFile(testApp.fullpath(tc.inFilename))
			if err != nil {
				t.Fatalf("Fail to read content for %s: %v", tc.inFilename, err)
			}

			if string(content) != tc.inContent {
				t.Errorf("Save note %s failed.\nGot : %v\nWant: %v", tc.inFilename, string(content), tc.inContent)
			}
		}

		// Check that original note is not modified
		if tc.outStatus != http.StatusOK {
			fi, err := os.Stat(path.Join(testApp.NotesDir, tc.inFilename))
			if err != nil {
				t.Fatalf("Fail to read content for %s: %#v", tc.inFilename, err)
			}
			if fi.IsDir() {
				continue
			}

			content, err := os.ReadFile(path.Join(testApp.NotesDir, tc.inFilename))
			if err != nil {
				t.Fatalf("Fail to read content for %s: %v", tc.inFilename, err)
			}

			if string(content) != testNotes[tc.inFilename] {
				t.Errorf("Save note %s modified original content.\nGot : %v\nWant: %v", tc.inFilename, string(content), testNotes[tc.inFilename])
			}
		}
	}
}

func setup(t *testing.T) (*WebApp, map[string]string) {
	testdir := t.TempDir()
	notesdir := path.Join(testdir, "notes")
	if err := os.Mkdir(notesdir, 0755); err != nil {
		t.Fatalf("fail to create test folder %s: %v", notesdir, err)
	}

	testCases := map[string]string{
		"../htpasswd":      "govid:$2b$10$ISdqfeODKyB4Qjd8KqA5BuP4whZY2bQlFmkrMoDhfLfyB1Xqx4c0Ov",
		"world_domination": "Use a giant magnet to attract Saturn to Earth.\nThe Brain.",
		"secret":           "Le roi Midas a des oreilles d'âne",
		"1.gif":            "GIF89a^A^@^A^@^@ÿ^@,^@^@^@^@^A^@^A^@^@^B^@;",
		"subdir/todo":      "Buy red socks",
	}

	for name, content := range testCases {
		filename := path.Join(notesdir, name)
		if err := os.MkdirAll(path.Dir(filename), 0750); err != nil {
			t.Fatalf("fail to create test environment for %s: %v", name, err)
		}

		if err := os.WriteFile(filename, []byte(content), 0660); err != nil {
			t.Fatalf("fail to create test environment for %s: %v", name, err)
		}
	}

	return NewWebApp(notesdir), testCases
}
