package internal

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestFetchMetadata(t *testing.T) {
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/api/mangas" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		if got := req.Header.Get("User-Agent"); got != "manga-converter" {
			t.Fatalf("unexpected user-agent: %s", got)
		}
		body := `[{"russian":"Боевая классика","url":"/mangas/42","image":{"original":"/covers/42.jpg"},"description":"Epic.","genres":["Action","Adventure"]}]`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	}))

	meta, err := FetchMetadata("Test Title")
	if err != nil {
		t.Fatalf("FetchMetadata error: %v", err)
	}

	if meta.Title != "Боевая классика" {
		t.Fatalf("unexpected title: %s", meta.Title)
	}
	if meta.URL != "https://shikimori.one/mangas/42" {
		t.Fatalf("unexpected url: %s", meta.URL)
	}
	if meta.Genres != "Action, Adventure" {
		t.Fatalf("unexpected genres: %s", meta.Genres)
	}
	if meta.CoverURL != "https://shikimori.one/covers/42.jpg" {
		t.Fatalf("unexpected cover url: %s", meta.CoverURL)
	}
}

func TestFetchMetadataNoResults(t *testing.T) {
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`[]`)),
			Header:     make(http.Header),
		}, nil
	}))

	if _, err := FetchMetadata("Unknown"); err == nil {
		t.Fatal("expected error when no results found")
	}
}

func TestFetchMetadataHTTPError(t *testing.T) {
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBuffer(nil)),
			Header:     make(http.Header),
		}, nil
	}))

	if _, err := FetchMetadata("ErrorCase"); err == nil {
		t.Fatal("expected error on non-200 status")
	}
}
