package storage

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadFromURLCreatesUniqueFilenames(t *testing.T) {
	store, err := NewLocalStorage(filepath.Join(t.TempDir(), "storage"), "http://localhost:5678/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write([]byte("fake-jpeg-content"))
	}))
	defer server.Close()

	firstURL, err := store.DownloadFromURL(server.URL+"/image", "images")
	if err != nil {
		t.Fatalf("first download failed: %v", err)
	}

	secondURL, err := store.DownloadFromURL(server.URL+"/image", "images")
	if err != nil {
		t.Fatalf("second download failed: %v", err)
	}

	if firstURL == secondURL {
		t.Fatalf("expected unique local urls, got %q", firstURL)
	}
	if !strings.HasSuffix(firstURL, ".jpg") || !strings.HasSuffix(secondURL, ".jpg") {
		t.Fatalf("expected jpeg extension, got %q and %q", firstURL, secondURL)
	}
}

func TestUploadCreatesUniqueFilenamesForSameOriginalName(t *testing.T) {
	store, err := NewLocalStorage(filepath.Join(t.TempDir(), "storage"), "http://localhost:5678/static")
	if err != nil {
		t.Fatalf("create local storage: %v", err)
	}

	firstURL, err := store.Upload(strings.NewReader("first"), "demo.png", "images")
	if err != nil {
		t.Fatalf("first upload failed: %v", err)
	}

	secondURL, err := store.Upload(strings.NewReader("second"), "demo.png", "images")
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}

	if firstURL == secondURL {
		t.Fatalf("expected unique upload urls, got %q", firstURL)
	}
}
