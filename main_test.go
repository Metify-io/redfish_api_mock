package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadAndValidateISO(t *testing.T) {
	image := make([]byte, 18*2048)
	copy(image[16*2048:], []byte{1, 'C', 'D', '0', '0', '1', 1})

	var receivedBytes int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "iso-user" || password != "iso-password" {
			t.Errorf("unexpected basic authentication: %q, %q, %v", username, password, ok)
		}
		receivedBytes, _ = w.Write(image)
	}))
	defer server.Close()

	if err := downloadAndValidateISO(context.Background(), server.URL+"/installer.iso", "iso-user", "iso-password"); err != nil {
		t.Fatalf("downloadAndValidateISO() error = %v", err)
	}
	if receivedBytes != len(image) {
		t.Fatalf("downloaded %d bytes, want %d", receivedBytes, len(image))
	}
}

func TestDownloadAndValidateISORejectsInvalidImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(make([]byte, 18*2048))
	}))
	defer server.Close()

	err := downloadAndValidateISO(context.Background(), server.URL+"/not-an-iso", "", "")
	if !errors.Is(err, errInvalidISO) {
		t.Fatalf("downloadAndValidateISO() error = %v, want errInvalidISO", err)
	}
}

func TestDownloadAndValidateISORejectsDownloadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	err := downloadAndValidateISO(context.Background(), server.URL+"/installer.iso", "", "")
	if err == nil || errors.Is(err, errInvalidISO) {
		t.Fatalf("downloadAndValidateISO() error = %v, want non-validation download error", err)
	}
}
