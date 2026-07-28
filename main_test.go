package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoadConfigAndConfiguredResponses(t *testing.T) {
	configFile, err := os.CreateTemp(t.TempDir(), "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = configFile.WriteString(`{
		"authentication": {"username": "bmc-user", "password": "bmc-secret"},
		"service_root": {"vendor": "Acme", "oem": {"Acme": {"Feature": "Enabled"}}},
		"system": {
			"manufacturer": "Acme", "model": "Rack 42", "installation_status_oem_key": "Acme",
			"oem": {"Acme": {"AssetTag": "lab-server"}}
		},
		"firmware_inventory": [
			{"id": "CPLD", "name": "System CPLD", "version": "4.2", "updateable": true, "software_id": "CPLD-4.2"}
		]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	loaded, err := loadConfig(configFile.Name())
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	previousConfig := config
	config = loaded
	t.Cleanup(func() { config = previousConfig })

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "1"}}
	getSystem(ctx)

	var system ComputerSystem
	if err := json.Unmarshal(recorder.Body.Bytes(), &system); err != nil {
		t.Fatalf("decode system response: %v", err)
	}
	if system.Manufacturer != "Acme" || system.Model != "Rack 42" {
		t.Fatalf("configured system identity = %q %q", system.Manufacturer, system.Model)
	}
	acme, ok := system.Oem["Acme"].(map[string]any)
	if !ok || acme["AssetTag"] != "lab-server" || acme["InstallationStatus"] != "Ready" {
		t.Fatalf("configured system OEM = %#v", system.Oem)
	}

	recorder = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "CPLD"}}
	getFirmwareInventoryItem(ctx)
	var firmware SoftwareInventory
	if err := json.Unmarshal(recorder.Body.Bytes(), &firmware); err != nil {
		t.Fatalf("decode firmware response: %v", err)
	}
	if firmware.ID != "CPLD" || firmware.Version != "4.2" {
		t.Fatalf("configured firmware = %#v", firmware)
	}

	router := gin.New()
	router.Use(basicAuth())
	router.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.SetBasicAuth("bmc-user", "bmc-secret")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("configured credentials status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

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
