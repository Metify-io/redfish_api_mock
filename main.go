package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type ServiceRoot struct {
	ODataContext string `json:"@odata.context"`
	ODataType    string `json:"@odata.type"`
	ODataID      string `json:"@odata.id"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	RedfishVersion string `json:"RedfishVersion"`
	UUID         string `json:"UUID"`
	Systems      Link   `json:"Systems"`
	Chassis      Link   `json:"Chassis"`
	Managers     Link   `json:"Managers"`
	SessionService Link `json:"SessionService"`
	UpdateService Link  `json:"UpdateService"`
}

type Link struct {
	ODataID string `json:"@odata.id"`
}

type Collection struct {
	ODataContext     string `json:"@odata.context"`
	ODataType        string `json:"@odata.type"`
	ODataID          string `json:"@odata.id"`
	Name             string `json:"Name"`
	MembersCount     int    `json:"Members@odata.count"`
	Members          []Link `json:"Members"`
}

type ComputerSystem struct {
	ODataContext    string `json:"@odata.context"`
	ODataType       string `json:"@odata.type"`
	ODataID         string `json:"@odata.id"`
	ID              string `json:"Id"`
	Name            string `json:"Name"`
	SystemType      string `json:"SystemType"`
	Manufacturer    string `json:"Manufacturer"`
	Model           string `json:"Model"`
	SerialNumber    string `json:"SerialNumber"`
	PartNumber      string `json:"PartNumber"`
	PowerState      string `json:"PowerState"`
	BiosVersion     string `json:"BiosVersion"`
	ProcessorSummary ProcessorSummary `json:"ProcessorSummary"`
	MemorySummary   MemorySummary    `json:"MemorySummary"`
	Status          Status           `json:"Status"`
}

type ProcessorSummary struct {
	Count int    `json:"Count"`
	Model string `json:"Model"`
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
	ODataContext     string `json:"@odata.context"`
	ODataType        string `json:"@odata.type"`
	ODataID          string `json:"@odata.id"`
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	ManagerType      string `json:"ManagerType"`
	FirmwareVersion  string `json:"FirmwareVersion"`
	Status           Status `json:"Status"`
}

type UpdateService struct {
	ODataContext      string     `json:"@odata.context"`
	ODataType         string     `json:"@odata.type"`
	ODataID           string     `json:"@odata.id"`
	ID                string     `json:"Id"`
	Name              string     `json:"Name"`
	ServiceEnabled    bool       `json:"ServiceEnabled"`
	HttpPushUri       string     `json:"HttpPushUri"`
	FirmwareInventory Link       `json:"FirmwareInventory"`
	Actions           UpdateServiceActions `json:"Actions"`
	Status            Status     `json:"Status"`
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
	ImageURI          string   `json:"ImageURI"`
	Targets           []string `json:"Targets,omitempty"`
	TransferProtocol  string   `json:"TransferProtocol,omitempty"`
	Username          string   `json:"Username,omitempty"`
	Password          string   `json:"Password,omitempty"`
	ForceUpdate       bool     `json:"ForceUpdate,omitempty"`
}

func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "admin" || password != "password" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Redfish"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("OData-Version", "4.0")
	json.NewEncoder(w).Encode(data)
}

func getServiceRoot(w http.ResponseWriter, r *http.Request) {
	serviceRoot := ServiceRoot{
		ODataContext:   "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
		ODataType:      "#ServiceRoot.v1_15_0.ServiceRoot",
		ODataID:        "/redfish/v1/",
		ID:             "RootService",
		Name:           "Root Service",
		RedfishVersion: "1.18.0",
		UUID:           "92384634-2938-2342-8820-489239905423",
		Systems:        Link{ODataID: "/redfish/v1/Systems"},
		Chassis:        Link{ODataID: "/redfish/v1/Chassis"},
		Managers:       Link{ODataID: "/redfish/v1/Managers"},
		SessionService: Link{ODataID: "/redfish/v1/SessionService"},
		UpdateService:  Link{ODataID: "/redfish/v1/UpdateService"},
	}
	jsonResponse(w, serviceRoot)
}

func getSystemsCollection(w http.ResponseWriter, r *http.Request) {
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
	jsonResponse(w, collection)
}

func getSystem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	systemID := vars["id"]
	
	system := ComputerSystem{
		ODataContext: "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
		ODataType:    "#ComputerSystem.v1_22_0.ComputerSystem",
		ODataID:      "/redfish/v1/Systems/" + systemID,
		ID:           systemID,
		Name:         "System",
		SystemType:   "Physical",
		Manufacturer: "Mock Vendor",
		Model:        "Mock Server X1000",
		SerialNumber: "MOCK123456789",
		PartNumber:   "MOCK-SRV-001",
		PowerState:   "On",
		BiosVersion:  "1.0.0",
		ProcessorSummary: ProcessorSummary{
			Count: 2,
			Model: "Mock CPU X5000",
			Status: Status{State: "Enabled", Health: "OK"},
		},
		MemorySummary: MemorySummary{
			TotalSystemMemoryGiB: 64,
			Status:               Status{State: "Enabled", Health: "OK"},
		},
		Status: Status{State: "Enabled", Health: "OK"},
	}
	jsonResponse(w, system)
}

func getChassisCollection(w http.ResponseWriter, r *http.Request) {
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
	jsonResponse(w, collection)
}

func getChassis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chassisID := vars["id"]
	
	chassis := Chassis{
		ODataContext: "/redfish/v1/$metadata#Chassis.Chassis",
		ODataType:    "#Chassis.v1_25_0.Chassis",
		ODataID:      "/redfish/v1/Chassis/" + chassisID,
		ID:           chassisID,
		Name:         "Chassis",
		ChassisType:  "RackMount",
		Manufacturer: "Mock Vendor",
		Model:        "Mock Chassis 1U",
		SerialNumber: "MOCK-CHASSIS-123",
		PartNumber:   "MOCK-CHS-001",
		Status:       Status{State: "Enabled", Health: "OK"},
	}
	jsonResponse(w, chassis)
}

func getManagersCollection(w http.ResponseWriter, r *http.Request) {
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
	jsonResponse(w, collection)
}

func getManager(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	managerID := vars["id"]
	
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
	jsonResponse(w, manager)
}

func getUpdateService(w http.ResponseWriter, r *http.Request) {
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
	jsonResponse(w, updateService)
}

func getFirmwareInventoryCollection(w http.ResponseWriter, r *http.Request) {
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
	jsonResponse(w, collection)
}

func getFirmwareInventoryItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["id"]
	
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
		http.NotFound(w, r)
		return
	}
	
	jsonResponse(w, inventory)
}

func simpleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req SimpleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.ImageURI == "" {
		http.Error(w, "ImageURI is required", http.StatusBadRequest)
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
	
	w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/1")
	w.WriteHeader(http.StatusAccepted)
	jsonResponse(w, response)
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()
	
	r := mux.NewRouter()
	
	// Apply basic auth to all routes
	r.Use(basicAuth)
	
	// Service root
	r.HandleFunc("/redfish/v1/", getServiceRoot).Methods("GET")
	r.HandleFunc("/redfish/v1", getServiceRoot).Methods("GET")
	
	// Systems endpoints
	r.HandleFunc("/redfish/v1/Systems", getSystemsCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Systems/", getSystemsCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Systems/{id}", getSystem).Methods("GET")
	
	// Chassis endpoints
	r.HandleFunc("/redfish/v1/Chassis", getChassisCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Chassis/", getChassisCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Chassis/{id}", getChassis).Methods("GET")
	
	// Managers endpoints
	r.HandleFunc("/redfish/v1/Managers", getManagersCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Managers/", getManagersCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/Managers/{id}", getManager).Methods("GET")
	
	// UpdateService endpoints
	r.HandleFunc("/redfish/v1/UpdateService", getUpdateService).Methods("GET")
	r.HandleFunc("/redfish/v1/UpdateService/", getUpdateService).Methods("GET")
	r.HandleFunc("/redfish/v1/UpdateService/FirmwareInventory", getFirmwareInventoryCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/UpdateService/FirmwareInventory/", getFirmwareInventoryCollection).Methods("GET")
	r.HandleFunc("/redfish/v1/UpdateService/FirmwareInventory/{id}", getFirmwareInventoryItem).Methods("GET")
	r.HandleFunc("/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate", simpleUpdate).Methods("POST")
	
	// Handle trailing slashes
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
			newPath := strings.TrimSuffix(r.URL.Path, "/")
			http.Redirect(w, r, newPath, http.StatusMovedPermanently)
			return
		}
		http.NotFound(w, r)
	})
	
	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Starting RedFish Mock Server on %s", addr)
	log.Println("Default credentials: admin / password")
	log.Fatal(http.ListenAndServe(addr, r))
}
