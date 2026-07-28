package main

type dellOEM struct{}

func (dellOEM) name() string { return "dell" }

func (dellOEM) resourceIDs() oemResourceIDs {
	return oemResourceIDs{
		System:       "System.Embedded.1",
		Chassis:      "System.Embedded.1",
		Manager:      "iDRAC.Embedded.1",
		VirtualMedia: "CD",
	}
}

func (dellOEM) applyDefaults(config *Config) {
	config.ServiceRoot.Product = "Integrated Dell Remote Access Controller"
	config.ServiceRoot.Vendor = "Dell Inc."
	config.ServiceRoot.Oem = map[string]any{
		"Dell": map[string]any{"@odata.type": "#DellServiceRoot.v1_0_0.DellServiceRoot"},
	}
	config.System.Manufacturer = "Dell Inc."
	config.System.Model = "PowerEdge"
	config.System.InstallationStatusOemKey = "Dell"
	config.Chassis.Manufacturer = "Dell Inc."
	config.Chassis.Model = "PowerEdge Chassis"
	config.Manager.Name = "iDRAC"
}
