package internal

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"One Piece", "One_Piece"},
		{"Attack_on_Titan", "Attack_on_Titan"},
		{" Leading and trailing spaces ", "_Leading_and_trailing_spaces_"},
	}

	for _, tc := range cases {
		if got := SafeName(tc.in); got != tc.want {
			t.Fatalf("SafeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsImage(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"cover.jpg", true},
		{"page.JPEG", true},
		{"scan.PNG", true},
		{"notes.txt", false},
		{"archive.zip", false},
	}

	for _, tc := range cases {
		if got := isImage(tc.name); got != tc.want {
			t.Fatalf("isImage(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestContainsImages(t *testing.T) {
	tmp := t.TempDir()

	withImages := filepath.Join(tmp, "with")
	withoutImages := filepath.Join(tmp, "without")

	writePNG(t, filepath.Join(withImages, "page.png"), 10, 10)
	if err := os.MkdirAll(withoutImages, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", withoutImages, err)
	}

	if !ContainsImages(withImages) {
		t.Fatal("ContainsImages should detect images")
	}
	if ContainsImages(withoutImages) {
		t.Fatal("ContainsImages should be false when no images present")
	}
}

func TestListImages(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "chapter", "001.png")
	b := filepath.Join(root, "002.jpeg")
	c := filepath.Join(root, "notes.txt")

	writePNG(t, a, 5, 5)
	writeJPEG(t, b, 5, 5)
	if err := os.WriteFile(c, []byte("not an image"), 0o644); err != nil {
		t.Fatalf("write %s: %v", c, err)
	}

	got, err := ListImages(root)
	if err != nil {
		t.Fatalf("ListImages error: %v", err)
	}

	want := []string{b, a}
	if len(got) != len(want) {
		t.Fatalf("ListImages returned %d files, want %d", len(got), len(want))
	}
	for i, path := range want {
		if got[i] != path {
			t.Fatalf("ListImages[%d] = %s, want %s", i, got[i], path)
		}
	}
}

func TestImageSize(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "sample.png")
	writePNG(t, tmp, 128, 256)

	w, h, err := ImageSize(tmp)
	if err != nil {
		t.Fatalf("ImageSize error: %v", err)
	}
	if w != 128 || h != 256 {
		t.Fatalf("ImageSize returned %dx%d, want 128x256", w, h)
	}
}

func TestUnzip(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	if err := os.MkdirAll(filepath.Join(source, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	filePath := filepath.Join(source, "nested", "file.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	zipPath := filepath.Join(dir, "archive.zip")
	zipDirectory(t, filepath.Join(source, "nested"), zipPath)

	dest := filepath.Join(dir, "out")
	if err := Unzip(zipPath, dest); err != nil {
		t.Fatalf("Unzip error: %v", err)
	}

	extracted := filepath.Join(dest, "nested", "file.txt")
	data, err := os.ReadFile(extracted)
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected content: %s", data)
	}
}

func TestUnzipPreservesStructure(t *testing.T) {
	dir := t.TempDir()
	var buffer bytes.Buffer
	zw := zip.NewWriter(&buffer)

	files := map[string]string{
		"root/file1.txt":    "first",
		"root/nested/file2": "second",
	}

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	zipPath := filepath.Join(dir, "structure.zip")
	if err := os.WriteFile(zipPath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip: %v", err)
	}

	dest := filepath.Join(dir, "out")
	if err := Unzip(zipPath, dest); err != nil {
		t.Fatalf("Unzip: %v", err)
	}

	for name, content := range files {
		path := filepath.Join(dest, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(data) != content {
			t.Fatalf("file %s = %q, want %q", path, data, content)
		}
	}
}
