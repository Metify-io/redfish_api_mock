package main

import (
	"flag"
	"log"
	"net/http"

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
}

type UpdateService struct {
	ODataContext      string               `json:"@odata.context"`
	ODataType         string               `json:"@odata.type"`
	ODataID           string               `json:"@odata.id"`
	ID                string               `json:"Id"`
	Name              string               `json:"Name"`
	ServiceEnabled    bool                 `json:"ServiceEnabled"`
	HttpPushUri       string               `json:"HttpPushUri"`
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
		UUID:           "92384634-2938-2342-8820-489239905423",
		Product:        "Mock RedFish Server v1.0",
		Vendor:         "Mock Vendor Corporation",
		Oem: map[string]interface{}{
			"Vendor": map[string]interface{}{
				"@odata.type":        "#MockVendorExtensions.v1_0_0.ServiceRoot",
				"ServerModel":        "Mock Enterprise Server X1000",
				"HardwareVersion":    "Rev 2.1",
				"ManagementVersion":  "BMC 3.2.1",
				"SupportContact":     "support@mockvendor.com",
				"WarrantyStatus":     "Active",
				"WarrantyExpiration": "2026-12-31",
			},
		},
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

	system := ComputerSystem{
		ODataContext: "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
		ODataType:    "#ComputerSystem.v1_22_0.ComputerSystem",
		ODataID:      "/redfish/v1/Systems/" + systemID,
		ID:           systemID,
		Name:         "System",
		SystemType:   "Physical",
		Manufacturer: "MetifyIO",
		Model:        "Mock Server X1000",
		SerialNumber: "MOCK123456789",
		PartNumber:   "MOCK-SRV-001",
		PowerState:   "On",
		BiosVersion:  "1.0.0",
		ProcessorSummary: ProcessorSummary{
			Count:  2,
			Model:  "Mock CPU X5000",
			Status: Status{State: "Enabled", Health: "OK"},
		},
		MemorySummary: MemorySummary{
			TotalSystemMemoryGiB: 64,
			Status:               Status{State: "Enabled", Health: "OK"},
		},
		Status: Status{State: "Enabled", Health: "OK"},
	}
	c.JSON(http.StatusOK, system)
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
		Name:         "Chassis",
		ChassisType:  "RackMount",
		Manufacturer: "Vendor",
		Model:        "Mock Chassis 1U",
		SerialNumber: "MOCK-CHASSIS-123",
		PartNumber:   "MOCK-CHS-001",
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
		Name:            "Manager",
		ManagerType:     "BMC",
		FirmwareVersion: "1.0.0",
		Status:          Status{State: "Enabled", Health: "OK"},
	}
	c.JSON(http.StatusOK, manager)
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
		HttpPushUri:       "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate",
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
	collection := Collection{
		ODataContext: "/redfish/v1/$metadata#SoftwareInventoryCollection.SoftwareInventoryCollection",
		ODataType:    "#SoftwareInventoryCollection.SoftwareInventoryCollection",
		ODataID:      "/redfish/v1/UpdateService/FirmwareInventory",
		Name:         "Firmware Inventory Collection",
		MembersCount: 3,
		Members: []Link{
			{ODataID: "/redfish/v1/UpdateService/FirmwareInventory/BIOS"},
			{ODataID: "/redfish/v1/UpdateService/FirmwareInventory/BMC"},
			{ODataID: "/redfish/v1/UpdateService/FirmwareInventory/NIC"},
		},
	}
	c.JSON(http.StatusOK, collection)
}

func getFirmwareInventoryItem(c *gin.Context) {
	c.Header("OData-Version", "4.0")
	itemID := c.Param("id")

	var inventory SoftwareInventory

	switch itemID {
	case "BIOS":
		inventory = SoftwareInventory{
			ODataContext: "/redfish/v1/$metadata#SoftwareInventory.SoftwareInventory",
			ODataType:    "#SoftwareInventory.v1_10_0.SoftwareInventory",
			ODataID:      "/redfish/v1/UpdateService/FirmwareInventory/BIOS",
			ID:           "BIOS",
			Name:         "System BIOS",
			Version:      "1.0.0",
			Updateable:   true,
			Status:       Status{State: "Enabled", Health: "OK"},
			SoftwareId:   "BIOS-1.0.0",
		}
	case "BMC":
		inventory = SoftwareInventory{
			ODataContext: "/redfish/v1/$metadata#SoftwareInventory.SoftwareInventory",
			ODataType:    "#SoftwareInventory.v1_10_0.SoftwareInventory",
			ODataID:      "/redfish/v1/UpdateService/FirmwareInventory/BMC",
			ID:           "BMC",
			Name:         "Baseboard Management Controller",
			Version:      "2.1.0",
			Updateable:   true,
			Status:       Status{State: "Enabled", Health: "OK"},
			SoftwareId:   "BMC-2.1.0",
		}
	case "NIC":
		inventory = SoftwareInventory{
			ODataContext: "/redfish/v1/$metadata#SoftwareInventory.SoftwareInventory",
			ODataType:    "#SoftwareInventory.v1_10_0.SoftwareInventory",
			ODataID:      "/redfish/v1/UpdateService/FirmwareInventory/NIC",
			ID:           "NIC",
			Name:         "Network Interface Controller",
			Version:      "3.2.1",
			Updateable:   true,
			Status:       Status{State: "Enabled", Health: "OK"},
			SoftwareId:   "NIC-3.2.1",
		}
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, inventory)
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
	response := map[string]interface{}{
		"@Message.ExtendedInfo": []map[string]interface{}{
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
	flag.Parse()

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

	// Chassis endpoints
	protected.GET("/Chassis", getChassisCollection)
	protected.GET("/Chassis/", getChassisCollection)
	protected.GET("/Chassis/:id", getChassis)

	// Manager individual endpoints (still protected)
	protected.GET("/Managers/:id", getManager)

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
