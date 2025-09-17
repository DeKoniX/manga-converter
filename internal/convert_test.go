package internal

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestIsZip(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"archive.zip", true},
		{"manga.cbz", false},
		{"no_extension", false},
		{"backup.zip.old", false},
	}

	for _, tc := range cases {
		if got := IsZip(tc.name); got != tc.want {
			t.Fatalf("IsZip(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestProcessZip(t *testing.T) {
	tmp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(cwd)
	})

	inputDir := filepath.Join(tmp, "input")
	workDir := filepath.Join(tmp, "workdir")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("mkdir input: %v", err)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	sourceRoot := filepath.Join(tmp, "source")
	volumeDir := filepath.Join(sourceRoot, "TestManga", "Volume 1")
	writeJPEG(t, filepath.Join(volumeDir, "page01.jpg"), 10, 10)

	zipPath := filepath.Join(inputDir, "test.zip")
	zipDirectory(t, filepath.Join(sourceRoot, "TestManga"), zipPath)

	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "shikimori.one" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if req.URL.Query().Get("search") != "TestManga" {
			t.Fatalf("unexpected search value: %s", req.URL.Query().Get("search"))
		}
		payload := []shikimoriResponse{{
			Russian:     "Test Title",
			URL:         "/mangas/1",
			Description: "Test description",
			Genres:      []string{"Action", "Drama"},
			Image: struct {
				Original string `json:"original"`
			}{Original: "/covers/1.jpg"},
		}}
		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	if err := ProcessZip("test.zip"); err != nil {
		t.Fatalf("ProcessZip error: %v", err)
	}

	if _, err := os.Stat(filepath.Join("input", "test.zip")); !os.IsNotExist(err) {
		t.Fatalf("input zip should be removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join("workdir", "test")); !os.IsNotExist(err) {
		t.Fatalf("workdir should be cleaned: %v", err)
	}

	cbzPath := filepath.Join("output", "cbz", "Test Title", "TestManga__Volume_1.cbz")
	if _, err := os.Stat(cbzPath); err != nil {
		t.Fatalf("expected CBZ at %s: %v", cbzPath, err)
	}

	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		t.Fatalf("open cbz: %v", err)
	}
	defer r.Close()

	var hasComicInfo bool
	for _, f := range r.File {
		if f.Name == "ComicInfo.xml" {
			hasComicInfo = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open ComicInfo: %v", err)
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("read ComicInfo: %v", err)
			}
			if !bytes.Contains(data, []byte("Test Title — Том Volume 1")) {
				t.Fatalf("ComicInfo.xml missing title, got %s", data)
			}
		}
	}
	if !hasComicInfo {
		t.Fatal("ComicInfo.xml not found in CBZ")
	}
}
