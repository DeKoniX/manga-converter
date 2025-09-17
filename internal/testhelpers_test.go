package internal

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func stubHTTPClient(t *testing.T, fn roundTripFunc) {
	t.Helper()
	original := http.DefaultClient.Transport
	http.DefaultClient.Transport = fn
	t.Cleanup(func() {
		http.DefaultClient.Transport = original
	})
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	t.Cleanup(func() { file.Close() })
	if err := png.Encode(file, img); err != nil {
		t.Fatalf("encode png %s: %v", path, err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close png %s: %v", path, err)
	}
}

func writeJPEG(t *testing.T, path string, width, height int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 80, G: 120, B: 200, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	t.Cleanup(func() { file.Close() })
	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatalf("encode jpeg %s: %v", path, err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close jpeg %s: %v", path, err)
	}
}

func zipDirectory(t *testing.T, rootDir, zipPath string) {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)

	base := filepath.Dir(rootDir)
	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		fw, err := writer.Create(rel)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, file); err != nil {
			file.Close()
			return err
		}
		return file.Close()
	}); err != nil {
		t.Fatalf("walk %s: %v", rootDir, err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		t.Fatalf("mkdir for zip %s: %v", zipPath, err)
	}
	if err := os.WriteFile(zipPath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("write zip %s: %v", zipPath, err)
	}
}
