package main

type ciscoOEM struct{}

func (ciscoOEM) name() string { return "cisco" }

func (ciscoOEM) resourceIDs() oemResourceIDs {
	return oemResourceIDs{System: "1", Chassis: "1", Manager: "CIMC", VirtualMedia: "CD"}
}

func (ciscoOEM) applyDefaults(config *Config) {
	config.ServiceRoot.Product = "Cisco Integrated Management Controller"
	config.ServiceRoot.Vendor = "Cisco Systems Inc."
	config.ServiceRoot.Oem = map[string]any{
		"Cisco": map[string]any{"@odata.type": "#CiscoServiceRootExtensions.v1_0_0.ServiceRoot"},
	}
	config.System.Manufacturer = "Cisco Systems Inc."
	config.System.Model = "UCS C-Series"
	config.System.InstallationStatusOemKey = "Cisco"
	config.Chassis.Manufacturer = "Cisco Systems Inc."
	config.Chassis.Model = "UCS C-Series Chassis"
	config.Manager.Name = "Cisco IMC"
}
