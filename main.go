package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceRoot struct {
	ODataContext   string                 `json:"@odata.context"`
	ODataType      string                 `json:"@odata.type"`
	ODataID        string                 `json:"@odata.id"`
	ID             string                 `json:"Id"`
	Name           string                 `json:"Name"`
	RedfishVersion string                 `json:"RedfishVersion"`
	UUID           string                 `json:"UUID"`
	Product        string                 `json:"Product"`
	Vendor         string                 `json:"Vendor"`
	Oem            map[string]interface{} `json:"Oem"`
	Systems        Link                   `json:"Systems"`
	Chassis        Link                   `json:"Chassis"`
	Managers       Link                   `json:"Managers"`
	SessionService Link                   `json:"SessionService"`
	UpdateService  Link                   `json:"UpdateService"`
	LicenseService Link                   `json:"LicenseService"`
}

type Link struct {
	ODataID string `json:"@odata.id"`
}

type Collection struct {
	ODataContext string `json:"@odata.context"`
	ODataType    string `json:"@odata.type"`
	ODataID      string `json:"@odata.id"`
	Name         string `json:"Name"`
	MembersCount int    `json:"Members@odata.count"`
	Members      []Link `json:"Members"`
}

type ComputerSystem struct {
	ODataContext     string           `json:"@odata.context"`
	ODataType        string           `json:"@odata.type"`
	ODataID          string           `json:"@odata.id"`
	ID               string           `json:"Id"`
	Name             string           `json:"Name"`
	SystemType       string           `json:"SystemType"`
	Manufacturer     string           `json:"Manufacturer"`
	Model            string           `json:"Model"`
	SerialNumber     string           `json:"SerialNumber"`
	PartNumber       string           `json:"PartNumber"`
	PowerState       string           `json:"PowerState"`
	BiosVersion      string           `json:"BiosVersion"`
	ProcessorSummary ProcessorSummary `json:"ProcessorSummary"`
	MemorySummary    MemorySummary    `json:"MemorySummary"`
	Status           Status           `json:"Status"`
	Boot             Boot             `json:"Boot"`
	Actions          SystemActions    `json:"Actions"`
	Oem              map[string]any   `json:"Oem"`
}

type Boot struct {
	BootSourceOverrideEnabled          string   `json:"BootSourceOverrideEnabled"`
	BootSourceOverrideTarget           string   `json:"BootSourceOverrideTarget"`
	BootSourceOverrideMode             string   `json:"BootSourceOverrideMode"`
	BootSourceOverrideEnabledAllowable []string `json:"BootSourceOverrideEnabled@Redfish.AllowableValues"`
	BootSourceOverrideTargetAllowable  []string `json:"BootSourceOverrideTarget@Redfish.AllowableValues"`
}

type SystemActions struct {
	Reset ResetAction `json:"#ComputerSystem.Reset"`
}

type ResetAction struct {
	Target          string   `json:"target"`
	AllowableValues []string `json:"ResetType@Redfish.AllowableValues"`
}

type ProcessorSummary struct {
	Count  int    `json:"Count"`
	Model  string `json:"Model"`
	Status Status `json:"Status"`
}

type MemorySummary struct {
	TotalSystemMemoryGiB int    `json:"TotalSystemMemoryGiB"`
	Status               Status `json:"Status"`
}

type Status struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}

type Config struct {
	ServiceRoot ServiceRootConfig    `json:"service_root"`
	System      SystemConfig         `json:"system"`
	Chassis     ChassisConfig        `json:"chassis"`
	Manager     ManagerConfig        `json:"manager"`
	Firmware    []FirmwareItemConfig `json:"firmware_inventory"`
}

type ServiceRootConfig struct {
	UUID    string         `json:"uuid"`
	Product string         `json:"product"`
	Vendor  string         `json:"vendor"`
	Oem     map[string]any `json:"oem"`
}

type SystemConfig struct {
	Name                     string         `json:"name"`
	SystemType               string         `json:"system_type"`
	Manufacturer             string         `json:"manufacturer"`
	Model                    string         `json:"model"`
	SerialNumber             string         `json:"serial_number"`
	PartNumber               string         `json:"part_number"`
	PowerState               string         `json:"power_state"`
	BiosVersion              string         `json:"bios_version"`
	ProcessorCount           int            `json:"processor_count"`
	ProcessorModel           string         `json:"processor_model"`
	TotalSystemMemoryGiB     int            `json:"total_system_memory_gib"`
	Oem                      map[string]any `json:"oem"`
	InstallationStatusOemKey string         `json:"installation_status_oem_key"`
}

type ChassisConfig struct {
	Name         string `json:"name"`
	ChassisType  string `json:"chassis_type"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	PartNumber   string `json:"part_number"`
}

type ManagerConfig struct {
	Name            string `json:"name"`
	ManagerType     string `json:"manager_type"`
	FirmwareVersion string `json:"firmware_version"`
}

type FirmwareItemConfig struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Updateable bool   `json:"updateable"`
	SoftwareID string `json:"software_id"`
}

var config = defaultConfig()

func defaultConfig() Config {
	return Config{
		ServiceRoot: ServiceRootConfig{
			UUID:    "92384634-2938-2342-8820-489239905423",
			Product: "Mock RedFish Server v1.0",
			Vendor:  "Mock Vendor Corporation",
			Oem: map[string]any{
				"Vendor": map[string]any{
					"@odata.type":        "#MockVendorExtensions.v1_0_0.ServiceRoot",
					"ServerModel":        "Mock Enterprise Server X1000",
					"HardwareVersion":    "Rev 2.1",
					"ManagementVersion":  "BMC 3.2.1",
					"SupportContact":     "support@mockvendor.com",
					"WarrantyStatus":     "Active",
					"WarrantyExpiration": "2026-12-31",
				},
			},
		},
		System: SystemConfig{
			Name:                     "System",
			SystemType:               "Physical",
			Manufacturer:             "MetifyIO",
			Model:                    "Mock Server X1000",
			SerialNumber:             "MOCK123456789",
			PartNumber:               "MOCK-SRV-001",
			PowerState:               "On",
			BiosVersion:              "1.0.0",
			ProcessorCount:           2,
			ProcessorModel:           "Mock CPU X5000",
			TotalSystemMemoryGiB:     64,
			Oem:                      map[string]any{},
			InstallationStatusOemKey: "MockVendor",
		},
		Chassis: ChassisConfig{
			Name:         "Chassis",
			ChassisType:  "RackMount",
			Manufacturer: "Vendor",
			Model:        "Mock Chassis 1U",
			SerialNumber: "MOCK-CHASSIS-123",
			PartNumber:   "MOCK-CHS-001",
		},
		Manager: ManagerConfig{
			Name:            "Manager",
			ManagerType:     "BMC",
			FirmwareVersion: "1.0.0",
		},
		Firmware: []FirmwareItemConfig{
			{ID: "BIOS", Name: "System BIOS", Version: "1.0.0", Updateable: true, SoftwareID: "BIOS-1.0.0"},
			{ID: "BMC", Name: "Baseboard Management Controller", Version: "2.1.0", Updateable: true, SoftwareID: "BMC-2.1.0"},
			{ID: "NIC", Name: "Network Interface Controller", Version: "3.2.1", Updateable: true, SoftwareID: "NIC-3.2.1"},
		},
	}
}

func loadConfig(path string) (Config, error) {
	loaded := defaultConfig()
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loaded); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	for i, item := range loaded.Firmware {
		if item.ID == "" {
			return Config{}, fmt.Errorf("firmware_inventory[%d].id is required", i)
		}
	}
	if loaded.System.InstallationStatusOemKey == "" {
		return Config{}, errors.New("system.installation_status_oem_key is required")
	}
	return loaded, nil
}

type Chassis struct {
	ODataContext string `json:"@odata.context"`
	ODataType    string `json:"@odata.type"`
	ODataID      string `json:"@odata.id"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	ChassisType  string `json:"ChassisType"`
	Manufacturer string `json:"Manufacturer"`
	Model        string `json:"Model"`
	SerialNumber string `json:"SerialNumber"`
	PartNumber   string `json:"PartNumber"`
	Status       Status `json:"Status"`
}

type Manager struct {
	ODataContext    string `json:"@odata.context"`
	ODataType       string `json:"@odata.type"`
	ODataID         string `json:"@odata.id"`
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	ManagerType     string `json:"ManagerType"`
	FirmwareVersion string `json:"FirmwareVersion"`
	Status          Status `json:"Status"`
	VirtualMedia    Link   `json:"VirtualMedia"`
}

type VirtualMedia struct {
	ODataContext   string              `json:"@odata.context"`
	ODataType      string              `json:"@odata.type"`
	ODataID        string              `json:"@odata.id"`
	ID             string              `json:"Id"`
	Name           string              `json:"Name"`
	MediaTypes     []string            `json:"MediaTypes"`
	Image          *string             `json:"Image"`
	Inserted       bool                `json:"Inserted"`
	ConnectedVia   string              `json:"ConnectedVia"`
	WriteProtected bool                `json:"WriteProtected"`
	Actions        VirtualMediaActions `json:"Actions"`
}

type VirtualMediaActions struct {
	InsertMedia VirtualMediaAction `json:"#VirtualMedia.InsertMedia"`
	EjectMedia  VirtualMediaAction `json:"#VirtualMedia.EjectMedia"`
}

type VirtualMediaAction struct {
	Target string `json:"target"`
}

type InsertMediaRequest struct {
	Image          string `json:"Image"`
	Inserted       *bool  `json:"Inserted,omitempty"`
	WriteProtected *bool  `json:"WriteProtected,omitempty"`
	UserName       string `json:"UserName,omitempty"`
	Password       string `json:"Password,omitempty"`
}

type SystemPatchRequest struct {
	Boot *struct {
		BootSourceOverrideEnabled *string `json:"BootSourceOverrideEnabled"`
		BootSourceOverrideTarget  *string `json:"BootSourceOverrideTarget"`
		BootSourceOverrideMode    *string `json:"BootSourceOverrideMode"`
	} `json:"Boot"`
}

type ResetRequest struct {
	ResetType string `json:"ResetType"`
}

type mockServerState struct {
	sync.RWMutex
	image                     string
	inserted                  bool
	writeProtected            bool
	bootSourceOverrideEnabled string
	bootSourceOverrideTarget  string
	bootSourceOverrideMode    string
	installationStatus        string
	installationStartedAt     time.Time
}

var mockState = mockServerState{
	writeProtected:            true,
	bootSourceOverrideEnabled: "Disabled",
	bootSourceOverrideTarget:  "None",
	bootSourceOverrideMode:    "UEFI",
	installationStatus:        "Ready",
}

var (
	errInvalidISO = errors.New("invalid ISO image")
	isoHTTPClient = &http.Client{Timeout: 30 * time.Minute}
)

type UpdateService struct {
	ODataContext      string               `json:"@odata.context"`
	ODataType         string               `json:"@odata.type"`
	ODataID           string               `json:"@odata.id"`
	ID                string               `json:"Id"`
	Name              string               `json:"Name"`
	ServiceEnabled    bool                 `json:"ServiceEnabled"`
	HTTPPushURI       string               `json:"HTTPPushURI"`
	FirmwareInventory Link                 `json:"FirmwareInventory"`
	Actions           UpdateServiceActions `json:"Actions"`
	Status            Status               `json:"Status"`
}

type UpdateServiceActions struct {
	SimpleUpdate UpdateServiceSimpleUpdate `json:"#UpdateService.SimpleUpdate"`
}

type UpdateServiceSimpleUpdate struct {
	Target string `json:"target"`
}

type SoftwareInventory struct {
	ODataContext string `json:"@odata.context"`
	ODataType    string `json:"@odata.type"`
	ODataID      string `json:"@odata.id"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Version      string `json:"Version"`
	Updateable   bool   `json:"Updateable"`
	Status       Status `json:"Status"`
	SoftwareId   string `json:"SoftwareId"`
}

type SimpleUpdateRequest struct {
	ImageURI         string   `json:"ImageURI"`
	Targets          []string `json:"Targets,omitempty"`
	TransferProtocol string   `json:"TransferProtocol,omitempty"`
	Username         string   `json:"Username,omitempty"`
	Password         string   `json:"Password,omitempty"`
	ForceUpdate      bool     `json:"ForceUpdate,omitempty"`
}

type LicenseService struct {
	ODataContext string `json:"@odata.context"`
	ODataType    string `json:"@odata.type"`
	ODataID      string `json:"@odata.id"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Licenses     Link   `json:"Licenses"`
	Status       Status `json:"Status"`
}

type License struct {
	ODataContext       string   `json:"@odata.context"`
	ODataType          string   `json:"@odata.type"`
	ODataID            string   `json:"@odata.id"`
	ID                 string   `json:"Id"`
	Name               string   `json:"Name"`
	LicenseType        string   `json:"LicenseType"`
	LicenseOrigin      string   `json:"LicenseOrigin"`
	ExpirationDate     string   `json:"ExpirationDate,omitempty"`
	InstallDate        string   `json:"InstallDate"`
	MaxAuthorizedCount int      `json:"MaxAuthorizedCount,omitempty"`
	RemainingUseCount  int      `json:"RemainingUseCount,omitempty"`
	Status             Status   `json:"Status"`
	Manufacturer       string   `json:"Manufacturer"`
	PartNumber         string   `json:"PartNumber,omitempty"`
	SerialNumber       string   `json:"SerialNumber,omitempty"`
	SKU                string   `json:"SKU,omitempty"`
	Links              struct{} `json:"Links"`
}

func basicAuth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		"admin": "password",
	})
}

func getServiceRoot(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	serviceRoot := ServiceRoot{
		ODataContext:   "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
		ODataType:      "#ServiceRoot.v1_15_0.ServiceRoot",
		ODataID:        "/redfish/v1/",
		ID:             "RootService",
		Name:           "Root Service",
		RedfishVersion: "1.18.0",
		UUID:           config.ServiceRoot.UUID,
		Product:        config.ServiceRoot.Product,
		Vendor:         config.ServiceRoot.Vendor,
		Oem:            config.ServiceRoot.Oem,
		Systems:        Link{ODataID: "/redfish/v1/Systems"},
		Chassis:        Link{ODataID: "/redfish/v1/Chassis"},
		Managers:       Link{ODataID: "/redfish/v1/Managers"},
		SessionService: Link{ODataID: "/redfish/v1/SessionService"},
		UpdateService:  Link{ODataID: "/redfish/v1/UpdateService"},
		LicenseService: Link{ODataID: "/redfish/v1/LicenseService"},
	}
	c.JSON(http.StatusOK, serviceRoot)
}

func getSystemsCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
		ODataType:    "#ComputerSystemCollection.ComputerSystemCollection",
		ODataID:      "/redfish/v1/Systems",
		Name:         "Computer System Collection",
		MembersCount: 1,
		Members: []Link{
			{ODataID: "/redfish/v1/Systems/1"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getSystem(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	systemID := c.Param("id")

	mockState.Lock()
	if mockState.installationStatus == "Installing" && time.Since(mockState.installationStartedAt) >= 2*time.Second {
		mockState.installationStatus = "Installed"
	}
	bootEnabled := mockState.bootSourceOverrideEnabled
	bootTarget := mockState.bootSourceOverrideTarget
	bootMode := mockState.bootSourceOverrideMode
	installationStatus := mockState.installationStatus
	mockState.Unlock()
	oem := make(map[string]any, len(config.System.Oem)+1)
	for key, value := range config.System.Oem {
		oem[key] = value
	}
	installationOem := map[string]any{}
	if configuredOem, ok := oem[config.System.InstallationStatusOemKey].(map[string]any); ok {
		for key, value := range configuredOem {
			installationOem[key] = value
		}
	}
	installationOem["InstallationStatus"] = installationStatus
	oem[config.System.InstallationStatusOemKey] = installationOem

	system := ComputerSystem{
		ODataContext: "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
		ODataType:    "#ComputerSystem.v1_22_0.ComputerSystem",
		ODataID:      "/redfish/v1/Systems/" + systemID,
		ID:           systemID,
		Name:         config.System.Name,
		SystemType:   config.System.SystemType,
		Manufacturer: config.System.Manufacturer,
		Model:        config.System.Model,
		SerialNumber: config.System.SerialNumber,
		PartNumber:   config.System.PartNumber,
		PowerState:   config.System.PowerState,
		BiosVersion:  config.System.BiosVersion,
		ProcessorSummary: ProcessorSummary{
			Count:  config.System.ProcessorCount,
			Model:  config.System.ProcessorModel,
			Status: Status{State: "Enabled", Health: "OK"},
		},
		MemorySummary: MemorySummary{
			TotalSystemMemoryGiB: config.System.TotalSystemMemoryGiB,
			Status:               Status{State: "Enabled", Health: "OK"},
		},
		Status: Status{State: "Enabled", Health: "OK"},
		Boot: Boot{
			BootSourceOverrideEnabled:          bootEnabled,
			BootSourceOverrideTarget:           bootTarget,
			BootSourceOverrideMode:             bootMode,
			BootSourceOverrideEnabledAllowable: []string{"Disabled", "Once", "Continuous"},
			BootSourceOverrideTargetAllowable:  []string{"None", "Cd", "Hdd", "Pxe", "Usb"},
		},
		Actions: SystemActions{
			Reset: ResetAction{
				Target:          "/redfish/v1/Systems/" + systemID + "/Actions/ComputerSystem.Reset",
				AllowableValues: []string{"On", "ForceOff", "GracefulShutdown", "GracefulRestart", "ForceRestart", "PowerCycle"},
			},
		},
		Oem: oem,
	}
	c.JSON(http.StatusOK, system)
}

func patchSystem(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	var req SystemPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Boot == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid Boot object is required"})
		return
	}

	mockState.Lock()
	defer mockState.Unlock()
	bootEnabled := mockState.bootSourceOverrideEnabled
	bootTarget := mockState.bootSourceOverrideTarget
	bootMode := mockState.bootSourceOverrideMode

	if req.Boot.BootSourceOverrideEnabled != nil {
		switch *req.Boot.BootSourceOverrideEnabled {
		case "Disabled", "Once", "Continuous":
			bootEnabled = *req.Boot.BootSourceOverrideEnabled
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported BootSourceOverrideEnabled value"})
			return
		}
	}
	if req.Boot.BootSourceOverrideTarget != nil {
		switch *req.Boot.BootSourceOverrideTarget {
		case "None", "Cd", "Hdd", "Pxe", "Usb":
			bootTarget = *req.Boot.BootSourceOverrideTarget
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported BootSourceOverrideTarget value"})
			return
		}
	}
	if req.Boot.BootSourceOverrideMode != nil {
		switch *req.Boot.BootSourceOverrideMode {
		case "UEFI", "Legacy":
			bootMode = *req.Boot.BootSourceOverrideMode
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported BootSourceOverrideMode value"})
			return
		}
	}
	mockState.bootSourceOverrideEnabled = bootEnabled
	mockState.bootSourceOverrideTarget = bootTarget
	mockState.bootSourceOverrideMode = bootMode

	c.Status(http.StatusNoContent)
}

func resetSystem(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	var req ResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	switch req.ResetType {
	case "On", "ForceOff", "GracefulShutdown", "GracefulRestart", "ForceRestart", "PowerCycle":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported ResetType value"})
		return
	}

	mockState.Lock()
	defer mockState.Unlock()
	bootsSystem := req.ResetType == "On" || req.ResetType == "GracefulRestart" || req.ResetType == "ForceRestart" || req.ResetType == "PowerCycle"
	if bootsSystem && mockState.inserted && mockState.bootSourceOverrideTarget == "Cd" && mockState.bootSourceOverrideEnabled != "Disabled" {
		mockState.installationStatus = "Installing"
		mockState.installationStartedAt = time.Now()
		if mockState.bootSourceOverrideEnabled == "Once" {
			mockState.bootSourceOverrideEnabled = "Disabled"
		}
	}

	c.Status(http.StatusNoContent)
}

func getChassisCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
		ODataType:    "#ChassisCollection.ChassisCollection",
		ODataID:      "/redfish/v1/Chassis",
		Name:         "Chassis Collection",
		MembersCount: 1,
		Members: []Link{
			{ODataID: "/redfish/v1/Chassis/1"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getChassis(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	chassisID := c.Param("id")

	chassis := Chassis{
		ODataContext: "/redfish/v1/$metadata#Chassis.Chassis",
		ODataType:    "#Chassis.v1_25_0.Chassis",
		ODataID:      "/redfish/v1/Chassis/" + chassisID,
		ID:           chassisID,
		Name:         config.Chassis.Name,
		ChassisType:  config.Chassis.ChassisType,
		Manufacturer: config.Chassis.Manufacturer,
		Model:        config.Chassis.Model,
		SerialNumber: config.Chassis.SerialNumber,
		PartNumber:   config.Chassis.PartNumber,
		Status:       Status{State: "Enabled", Health: "OK"},
	}
	c.JSON(http.StatusOK, chassis)
}

func getManagersCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
		ODataType:    "#ManagerCollection.ManagerCollection",
		ODataID:      "/redfish/v1/Managers",
		Name:         "Manager Collection",
		MembersCount: 1,
		Members: []Link{
			{ODataID: "/redfish/v1/Managers/1"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getManager(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	managerID := c.Param("id")

	manager := Manager{
		ODataContext:    "/redfish/v1/$metadata#Manager.Manager",
		ODataType:       "#Manager.v1_19_0.Manager",
		ODataID:         "/redfish/v1/Managers/" + managerID,
		ID:              managerID,
		Name:            config.Manager.Name,
		ManagerType:     config.Manager.ManagerType,
		FirmwareVersion: config.Manager.FirmwareVersion,
		Status:          Status{State: "Enabled", Health: "OK"},
		VirtualMedia:    Link{ODataID: "/redfish/v1/Managers/" + managerID + "/VirtualMedia"},
	}
	c.JSON(http.StatusOK, manager)
}

func getVirtualMediaCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	managerID := c.Param("id")
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#VirtualMediaCollection.VirtualMediaCollection",
		ODataType:    "#VirtualMediaCollection.VirtualMediaCollection",
		ODataID:      "/redfish/v1/Managers/" + managerID + "/VirtualMedia",
		Name:         "Virtual Media Collection",
		MembersCount: 1,
		Members: []Link{
			{ODataID: "/redfish/v1/Managers/" + managerID + "/VirtualMedia/CD"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getVirtualMedia(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	if c.Param("mediaID") != "CD" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Virtual media device not found"})
		return
	}

	managerID := c.Param("id")
	baseURI := "/redfish/v1/Managers/" + managerID + "/VirtualMedia/CD"
	mockState.RLock()
	var image *string
	if mockState.image != "" {
		imageValue := mockState.image
		image = &imageValue
	}
	inserted := mockState.inserted
	writeProtected := mockState.writeProtected
	mockState.RUnlock()

	connectedVia := "NotConnected"
	if inserted {
		connectedVia = "URI"
	}
	media := VirtualMedia{
		ODataContext:   "/redfish/v1/$metadata#VirtualMedia.VirtualMedia",
		ODataType:      "#VirtualMedia.v1_6_0.VirtualMedia",
		ODataID:        baseURI,
		ID:             "CD",
		Name:           "Virtual CD/DVD",
		MediaTypes:     []string{"CD", "DVD"},
		Image:          image,
		Inserted:       inserted,
		ConnectedVia:   connectedVia,
		WriteProtected: writeProtected,
		Actions: VirtualMediaActions{
			InsertMedia: VirtualMediaAction{Target: baseURI + "/Actions/VirtualMedia.InsertMedia"},
			EjectMedia:  VirtualMediaAction{Target: baseURI + "/Actions/VirtualMedia.EjectMedia"},
		},
	}
	c.JSON(http.StatusOK, media)
}

func downloadAndValidateISO(ctx context.Context, imageURL, username, password string) error {
	parsedURL, err := url.ParseRequestURI(imageURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return fmt.Errorf("%w: Image must be an HTTP or HTTPS URL", errInvalidISO)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("create image request: %w", err)
	}
	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	response, err := isoHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download image: server returned %s", response.Status)
	}

	image, err := os.CreateTemp("", "redfish-virtual-media-*.iso")
	if err != nil {
		return fmt.Errorf("create temporary image: %w", err)
	}
	defer os.Remove(image.Name())
	defer image.Close()

	size, err := io.Copy(image, response.Body)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}
	if err := validateISO(image, size); err != nil {
		return err
	}
	return nil
}

func validateISO(image *os.File, size int64) error {
	const (
		sectorSize            = int64(2048)
		firstDescriptorSector = int64(16)
		primaryDescriptor     = byte(1)
		terminatingDescriptor = byte(255)
	)

	header := make([]byte, 7)
	for sector := firstDescriptorSector; sector*sectorSize+int64(len(header)) <= size; sector++ {
		if _, err := image.ReadAt(header, sector*sectorSize); err != nil {
			return fmt.Errorf("%w: cannot read volume descriptor: %v", errInvalidISO, err)
		}
		if !bytes.Equal(header[1:6], []byte("CD001")) || header[6] != 1 {
			return fmt.Errorf("%w: ISO-9660 volume descriptor not found", errInvalidISO)
		}
		switch header[0] {
		case primaryDescriptor:
			return nil
		case terminatingDescriptor:
			return fmt.Errorf("%w: primary volume descriptor not found", errInvalidISO)
		}
	}
	return fmt.Errorf("%w: image is too small or has no primary volume descriptor", errInvalidISO)
}

func insertMedia(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	if c.Param("mediaID") != "CD" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Virtual media device not found"})
		return
	}

	var req InsertMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	inserted := true
	if req.Inserted != nil {
		inserted = *req.Inserted
	}
	writeProtected := true
	if req.WriteProtected != nil {
		writeProtected = *req.WriteProtected
	}
	if err := downloadAndValidateISO(c.Request.Context(), req.Image, req.UserName, req.Password); err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, errInvalidISO) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	mockState.Lock()
	mockState.image = req.Image
	mockState.inserted = inserted
	mockState.writeProtected = writeProtected
	if inserted {
		mockState.installationStatus = "MediaMounted"
	}
	mockState.Unlock()

	c.Status(http.StatusNoContent)
}

func ejectMedia(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	if c.Param("mediaID") != "CD" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Virtual media device not found"})
		return
	}

	mockState.Lock()
	mockState.image = ""
	mockState.inserted = false
	mockState.writeProtected = true
	if mockState.installationStatus != "Installed" {
		mockState.installationStatus = "Ready"
	}
	mockState.Unlock()

	c.Status(http.StatusNoContent)
}

func getUpdateService(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	updateService := UpdateService{
		ODataContext:      "/redfish/v1/$metadata#UpdateService.UpdateService",
		ODataType:         "#UpdateService.v1_12_0.UpdateService",
		ODataID:           "/redfish/v1/UpdateService",
		ID:                "UpdateService",
		Name:              "Update Service",
		ServiceEnabled:    true,
		HTTPPushURI:       "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate",
		FirmwareInventory: Link{ODataID: "/redfish/v1/UpdateService/FirmwareInventory"},
		Actions: UpdateServiceActions{
			SimpleUpdate: UpdateServiceSimpleUpdate{
				Target: "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate",
			},
		},
		Status: Status{State: "Enabled", Health: "OK"},
	}
	c.JSON(http.StatusOK, updateService)
}

func getFirmwareInventoryCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	members := make([]Link, 0, len(config.Firmware))
	for _, item := range config.Firmware {
		members = append(members, Link{ODataID: "/redfish/v1/UpdateService/FirmwareInventory/" + item.ID})
	}
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#SoftwareInventoryCollection.SoftwareInventoryCollection",
		ODataType:    "#SoftwareInventoryCollection.SoftwareInventoryCollection",
		ODataID:      "/redfish/v1/UpdateService/FirmwareInventory",
		Name:         "Firmware Inventory Collection",
		MembersCount: len(members),
		Members:      members,
	}
	c.JSON(http.StatusOK, collection)
}

func getFirmwareInventoryItem(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	itemID := c.Param("id")

	for _, item := range config.Firmware {
		if item.ID == itemID {
			c.JSON(http.StatusOK, SoftwareInventory{
				ODataContext: "/redfish/v1/$metadata#SoftwareInventory.SoftwareInventory",
				ODataType:    "#SoftwareInventory.v1_10_0.SoftwareInventory",
				ODataID:      "/redfish/v1/UpdateService/FirmwareInventory/" + item.ID,
				ID:           item.ID,
				Name:         item.Name,
				Version:      item.Version,
				Updateable:   item.Updateable,
				Status:       Status{State: "Enabled", Health: "OK"},
				SoftwareId:   item.SoftwareID,
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}

func simpleUpdate(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	var req SimpleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if req.ImageURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ImageURI is required"})
		return
	}

	// Mock response - in a real implementation, this would start an update task
	response := map[string]any{
		"@Message.ExtendedInfo": []map[string]any{
			{
				"MessageId": "Update.1.0.0.UpdateInProgress",
				"Message":   "The update operation has been started and is in progress.",
				"Severity":  "OK",
			},
		},
	}

	c.Header("Location", "/redfish/v1/TaskService/Tasks/1")
	c.JSON(http.StatusAccepted, response)
}

func getLicenseService(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	licenseService := LicenseService{
		ODataContext: "/redfish/v1/$metadata#LicenseService.LicenseService",
		ODataType:    "#LicenseService.v1_1_0.LicenseService",
		ODataID:      "/redfish/v1/LicenseService",
		ID:           "LicenseService",
		Name:         "License Service",
		Licenses:     Link{ODataID: "/redfish/v1/LicenseService/Licenses"},
		Status:       Status{State: "Enabled", Health: "OK"},
	}
	c.JSON(http.StatusOK, licenseService)
}

func getLicensesCollection(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#LicenseCollection.LicenseCollection",
		ODataType:    "#LicenseCollection.LicenseCollection",
		ODataID:      "/redfish/v1/LicenseService/Licenses",
		Name:         "License Collection",
		MembersCount: 2,
		Members: []Link{
			{ODataID: "/redfish/v1/LicenseService/Licenses/BMC-License"},
			{ODataID: "/redfish/v1/LicenseService/Licenses/BIOS-License"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getLicense(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	licenseID := c.Param("id")

	var license License

	switch licenseID {
	case "BMC-License":
		license = License{
			ODataContext:       "/redfish/v1/$metadata#License.License",
			ODataType:          "#License.v1_1_0.License",
			ODataID:            "/redfish/v1/LicenseService/Licenses/BMC-License",
			ID:                 "BMC-License",
			Name:               "BMC Management License",
			LicenseType:        "Production",
			LicenseOrigin:      "BuiltIn",
			InstallDate:        "2024-01-15T08:00:00Z",
			ExpirationDate:     "2026-01-15T08:00:00Z",
			MaxAuthorizedCount: 1,
			RemainingUseCount:  1,
			Status:             Status{State: "Enabled", Health: "OK"},
			Manufacturer:       "Mock Vendor Corporation",
			PartNumber:         "BMC-LIC-001",
			SerialNumber:       "BMC123456789",
			SKU:                "BMC-PROD-LIC",
			Links:              struct{}{},
		}
	case "BIOS-License":
		license = License{
			ODataContext:  "/redfish/v1/$metadata#License.License",
			ODataType:     "#License.v1_1_0.License",
			ODataID:       "/redfish/v1/LicenseService/Licenses/BIOS-License",
			ID:            "BIOS-License",
			Name:          "BIOS Feature License",
			LicenseType:   "Production",
			LicenseOrigin: "BuiltIn",
			InstallDate:   "2024-01-15T08:00:00Z",
			Status:        Status{State: "Enabled", Health: "OK"},
			Manufacturer:  "Mock Vendor Corporation",
			PartNumber:    "BIOS-LIC-001",
			SerialNumber:  "BIOS123456789",
			SKU:           "BIOS-PROD-LIC",
			Links:         struct{}{},
		}
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, license)
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	host := flag.String("host", "localhost", "Host to listen on")
	configPath := flag.String("config", "config.json", "Path to mock data config file")
	flag.Parse()

	loadedConfig, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config %q: %v", *configPath, err)
	}
	config = loadedConfig

	r := gin.Default()

	// Public endpoints (no auth required)
	r.GET("/redfish/v1/", getServiceRoot)
	r.GET("/redfish/v1", getServiceRoot)
	r.GET("/redfish/v1/Managers", getManagersCollection)
	r.GET("/redfish/v1/Managers/", getManagersCollection)

	// Protected endpoints (require auth)
	protected := r.Group("/redfish/v1")
	protected.Use(basicAuth())

	// Systems endpoints
	protected.GET("/Systems", getSystemsCollection)
	protected.GET("/Systems/", getSystemsCollection)
	protected.GET("/Systems/:id", getSystem)
	protected.PATCH("/Systems/:id", patchSystem)
	protected.POST("/Systems/:id/Actions/ComputerSystem.Reset", resetSystem)

	// Chassis endpoints
	protected.GET("/Chassis", getChassisCollection)
	protected.GET("/Chassis/", getChassisCollection)
	protected.GET("/Chassis/:id", getChassis)

	// Manager individual endpoints (still protected)
	protected.GET("/Managers/:id", getManager)
	protected.GET("/Managers/:id/VirtualMedia", getVirtualMediaCollection)
	protected.GET("/Managers/:id/VirtualMedia/", getVirtualMediaCollection)
	protected.GET("/Managers/:id/VirtualMedia/:mediaID", getVirtualMedia)
	protected.POST("/Managers/:id/VirtualMedia/:mediaID/Actions/VirtualMedia.InsertMedia", insertMedia)
	protected.POST("/Managers/:id/VirtualMedia/:mediaID/Actions/VirtualMedia.EjectMedia", ejectMedia)

	// UpdateService endpoints
	protected.GET("/UpdateService", getUpdateService)
	protected.GET("/UpdateService/", getUpdateService)
	protected.GET("/UpdateService/FirmwareInventory", getFirmwareInventoryCollection)
	protected.GET("/UpdateService/FirmwareInventory/", getFirmwareInventoryCollection)
	protected.GET("/UpdateService/FirmwareInventory/:id", getFirmwareInventoryItem)
	protected.POST("/UpdateService/Actions/UpdateService.SimpleUpdate", simpleUpdate)

	// LicenseService endpoints
	protected.GET("/LicenseService", getLicenseService)
	protected.GET("/LicenseService/", getLicenseService)
	protected.GET("/LicenseService/Licenses", getLicensesCollection)
	protected.GET("/LicenseService/Licenses/", getLicensesCollection)
	protected.GET("/LicenseService/Licenses/:id", getLicense)

	addr := *host + ":" + *port
	log.Printf("\nStarting RedFish Mock Server on %s", addr)
	log.Println("\nDefault credentials: admin / password")
	log.Fatal(r.Run(addr))
}
