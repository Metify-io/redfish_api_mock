package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOEMProfiles(t *testing.T) {
	tests := []struct {
		name           string
		vendor         string
		manufacturer   string
		systemID       string
		managerID      string
		virtualMediaID string
		oemKey         string
	}{
		{name: "mock", vendor: "Mock Vendor Corporation", manufacturer: "MetifyIO", systemID: "1", managerID: "1", virtualMediaID: "CD", oemKey: "MockVendor"},
		{name: "supermicro", vendor: "Supermicro", manufacturer: "Supermicro", systemID: "1", managerID: "1", virtualMediaID: "CD1", oemKey: "Supermicro"},
		{name: "dell", vendor: "Dell Inc.", manufacturer: "Dell Inc.", systemID: "System.Embedded.1", managerID: "iDRAC.Embedded.1", virtualMediaID: "CD", oemKey: "Dell"},
		{name: "cisco", vendor: "Cisco Systems Inc.", manufacturer: "Cisco Systems Inc.", systemID: "1", managerID: "CIMC", virtualMediaID: "CD", oemKey: "Cisco"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(`{"oem":"`+test.name+`"}`), 0o600); err != nil {
				t.Fatal(err)
			}
			loaded, err := loadConfig(path)
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}
			if loaded.ServiceRoot.Vendor != test.vendor || loaded.System.Manufacturer != test.manufacturer {
				t.Fatalf("profile identity = %q, %q", loaded.ServiceRoot.Vendor, loaded.System.Manufacturer)
			}
			if loaded.System.InstallationStatusOemKey != test.oemKey {
				t.Fatalf("installation OEM key = %q, want %q", loaded.System.InstallationStatusOemKey, test.oemKey)
			}

			behavior, err := oemBehaviorFor(loaded.OEM)
			if err != nil {
				t.Fatal(err)
			}
			ids := behavior.resourceIDs()
			if ids.System != test.systemID || ids.Manager != test.managerID || ids.VirtualMedia != test.virtualMediaID {
				t.Fatalf("resource IDs = %#v", ids)
			}
		})
	}
}

func TestSupermicroProvisioningBootOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"oem":"supermicro"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	previousConfig := config
	config = loaded
	mockState.Lock()
	previousImage := mockState.image
	previousInserted := mockState.inserted
	previousWriteProtected := mockState.writeProtected
	previousBootEnabled := mockState.bootSourceOverrideEnabled
	previousBootTarget := mockState.bootSourceOverrideTarget
	previousBootMode := mockState.bootSourceOverrideMode
	previousInstallationStatus := mockState.installationStatus
	previousInstallationStartedAt := mockState.installationStartedAt
	mockState.bootSourceOverrideEnabled = "Disabled"
	mockState.bootSourceOverrideTarget = "None"
	mockState.bootSourceOverrideMode = "UEFI"
	mockState.Unlock()
	t.Cleanup(func() {
		config = previousConfig
		mockState.Lock()
		mockState.image = previousImage
		mockState.inserted = previousInserted
		mockState.writeProtected = previousWriteProtected
		mockState.bootSourceOverrideEnabled = previousBootEnabled
		mockState.bootSourceOverrideTarget = previousBootTarget
		mockState.bootSourceOverrideMode = previousBootMode
		mockState.installationStatus = previousInstallationStatus
		mockState.installationStartedAt = previousInstallationStartedAt
		mockState.Unlock()
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PATCH("/redfish/v1/Systems/:id", patchSystem)
	router.GET("/redfish/v1/Systems/:id", getSystem)

	payload := []byte(`{
  "Boot": {
    "BootSourceOverrideEnabled": "Once",
    "BootSourceOverrideMode": "Legacy",
    "BootSourceOverrideTarget": "UsbCd"
  }
}`)
	request := httptest.NewRequest(http.MethodPatch, "/redfish/v1/Systems/1", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("PATCH status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var patchedSystem ComputerSystem
	if err := json.Unmarshal(recorder.Body.Bytes(), &patchedSystem); err != nil {
		t.Fatalf("decode PATCH response: %v", err)
	}
	if patchedSystem.Boot.BootSourceOverrideEnabled != "Once" || patchedSystem.Boot.BootSourceOverrideMode != "Legacy" || patchedSystem.Boot.BootSourceOverrideTarget != "UsbCd" {
		t.Fatalf("PATCH response boot override = %#v", patchedSystem.Boot)
	}

	request = httptest.NewRequest(http.MethodGet, "/redfish/v1/Systems/1", nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var system ComputerSystem
	if err := json.Unmarshal(recorder.Body.Bytes(), &system); err != nil {
		t.Fatalf("decode system response: %v", err)
	}
	if system.Boot.BootSourceOverrideEnabled != "Once" || system.Boot.BootSourceOverrideMode != "Legacy" || system.Boot.BootSourceOverrideTarget != "UsbCd" {
		t.Fatalf("persisted boot override = %#v", system.Boot)
	}
	for _, target := range system.Boot.BootSourceOverrideTargetAllowable {
		if target == "UsbCd" {
			return
		}
	}
	t.Fatalf("allowable boot targets = %v, want UsbCd", system.Boot.BootSourceOverrideTargetAllowable)
}

func TestResetSystemUpdatesPowerState(t *testing.T) {
	previousConfig := config
	config = defaultConfig()
	mockState.Lock()
	previousPowerState := mockState.powerState
	mockState.Unlock()
	t.Cleanup(func() {
		config = previousConfig
		mockState.Lock()
		mockState.powerState = previousPowerState
		mockState.Unlock()
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/redfish/v1/Systems/:id/Actions/ComputerSystem.Reset", resetSystem)
	router.GET("/redfish/v1/Systems/:id", getSystem)

	tests := []struct {
		resetType         string
		initialPowerState string
		wantPowerState    string
	}{
		{resetType: "ForceOff", initialPowerState: "On", wantPowerState: "Off"},
		{resetType: "GracefulShutdown", initialPowerState: "On", wantPowerState: "Off"},
		{resetType: "On", initialPowerState: "Off", wantPowerState: "On"},
		{resetType: "GracefulRestart", initialPowerState: "Off", wantPowerState: "On"},
		{resetType: "ForceRestart", initialPowerState: "Off", wantPowerState: "On"},
		{resetType: "PowerCycle", initialPowerState: "Off", wantPowerState: "On"},
	}

	for _, test := range tests {
		t.Run(test.resetType, func(t *testing.T) {
			mockState.Lock()
			mockState.powerState = test.initialPowerState
			mockState.Unlock()

			payload := []byte(`{"ResetType":"` + test.resetType + `"}`)
			request := httptest.NewRequest(http.MethodPost, "/redfish/v1/Systems/1/Actions/ComputerSystem.Reset", bytes.NewReader(payload))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusNoContent {
				t.Fatalf("POST status = %d, want %d; body = %s", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}

			request = httptest.NewRequest(http.MethodGet, "/redfish/v1/Systems/1", nil)
			recorder = httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET status = %d, want %d", recorder.Code, http.StatusOK)
			}
			var system ComputerSystem
			if err := json.Unmarshal(recorder.Body.Bytes(), &system); err != nil {
				t.Fatalf("decode system response: %v", err)
			}
			if system.PowerState != test.wantPowerState {
				t.Fatalf("PowerState = %q, want %q", system.PowerState, test.wantPowerState)
			}
		})
	}
}

func TestConfiguredEthernetInterface(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{
  "ethernet_interface": {
    "id": "PXE1",
    "name": "Provisioning NIC",
    "mac_address": "02:12:34:56:78:9A"
  }
}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadConfig(path)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	previousConfig := config
	config = loaded
	t.Cleanup(func() { config = previousConfig })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/redfish/v1/Systems/:id", getSystem)
	router.GET("/redfish/v1/Systems/:id/EthernetInterfaces", getEthernetInterfacesCollection)
	router.GET("/redfish/v1/Systems/:id/EthernetInterfaces/:interfaceID", getEthernetInterface)

	request := httptest.NewRequest(http.MethodGet, "/redfish/v1/Systems/1", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("system GET status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var system ComputerSystem
	if err := json.Unmarshal(recorder.Body.Bytes(), &system); err != nil {
		t.Fatalf("decode system response: %v", err)
	}
	if system.EthernetInterfaces.ODataID != "/redfish/v1/Systems/1/EthernetInterfaces" {
		t.Fatalf("EthernetInterfaces link = %q", system.EthernetInterfaces.ODataID)
	}

	request = httptest.NewRequest(http.MethodGet, system.EthernetInterfaces.ODataID, nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("collection GET status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var collection Collection
	if err := json.Unmarshal(recorder.Body.Bytes(), &collection); err != nil {
		t.Fatalf("decode Ethernet interface collection: %v", err)
	}
	if collection.MembersCount != 1 || len(collection.Members) != 1 {
		t.Fatalf("Ethernet interface collection = %#v", collection)
	}

	request = httptest.NewRequest(http.MethodGet, collection.Members[0].ODataID, nil)
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("interface GET status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var ethernetInterface EthernetInterface
	if err := json.Unmarshal(recorder.Body.Bytes(), &ethernetInterface); err != nil {
		t.Fatalf("decode Ethernet interface: %v", err)
	}
	if ethernetInterface.ID != "PXE1" || ethernetInterface.Name != "Provisioning NIC" {
		t.Fatalf("Ethernet interface identity = %#v", ethernetInterface)
	}
	if ethernetInterface.MACAddress != "02:12:34:56:78:9A" || ethernetInterface.PermanentMACAddress != "02:12:34:56:78:9A" {
		t.Fatalf("Ethernet interface MAC addresses = %q, %q", ethernetInterface.MACAddress, ethernetInterface.PermanentMACAddress)
	}
}

func TestLoadConfigRejectsUnsupportedOEM(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"oem":"unknown"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConfig(path); err == nil {
		t.Fatal("loadConfig() succeeded for unsupported OEM")
	}
}

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
