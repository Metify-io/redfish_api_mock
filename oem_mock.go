package main

type mockOEM struct{}

func (mockOEM) name() string { return "mock" }

func (mockOEM) resourceIDs() oemResourceIDs {
	return oemResourceIDs{System: "1", Chassis: "1", Manager: "1", VirtualMedia: "CD"}
}

func (mockOEM) applyDefaults(config *Config) {
	config.ServiceRoot.Product = "Mock RedFish Server v1.0"
	config.ServiceRoot.Vendor = "Mock Vendor Corporation"
	config.ServiceRoot.Oem = map[string]any{
		"Vendor": map[string]any{
			"@odata.type":        "#MockVendorExtensions.v1_0_0.ServiceRoot",
			"ServerModel":        "Mock Enterprise Server X1000",
			"HardwareVersion":    "Rev 2.1",
			"ManagementVersion":  "BMC 3.2.1",
			"SupportContact":     "support@mockvendor.com",
			"WarrantyStatus":     "Active",
			"WarrantyExpiration": "2026-12-31",
		},
	}
	config.System.Manufacturer = "MetifyIO"
	config.System.Model = "Mock Server X1000"
	config.System.InstallationStatusOemKey = "MockVendor"
	config.Chassis.Manufacturer = "Vendor"
	config.Chassis.Model = "Mock Chassis 1U"
	config.Manager.Name = "Manager"
}
