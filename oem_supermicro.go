package main

type supermicroOEM struct{}

func (supermicroOEM) name() string { return "supermicro" }

func (supermicroOEM) resourceIDs() oemResourceIDs {
	return oemResourceIDs{System: "1", Chassis: "1", Manager: "1", VirtualMedia: "CD1"}
}

func (supermicroOEM) applyDefaults(config *Config) {
	config.ServiceRoot.Product = "Supermicro Redfish Service"
	config.ServiceRoot.Vendor = "Supermicro"
	config.ServiceRoot.Oem = map[string]any{
		"Supermicro": map[string]any{"@odata.type": "#SmcServiceRootExtensions.v1_0_0.ServiceRoot"},
	}
	config.System.Manufacturer = "Supermicro"
	config.System.Model = "SuperServer"
	config.System.InstallationStatusOemKey = "Supermicro"
	config.Chassis.Manufacturer = "Supermicro"
	config.Chassis.Model = "SuperServer Chassis"
	config.Manager.Name = "BMC"
}
