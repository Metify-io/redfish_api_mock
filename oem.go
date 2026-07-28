package main

import (
	"fmt"
	"strings"
)

type oemResourceIDs struct {
	System       string
	Chassis      string
	Manager      string
	VirtualMedia string
}

type oemBehavior interface {
	name() string
	applyDefaults(*Config)
	resourceIDs() oemResourceIDs
}

func oemBehaviorFor(name string) (oemBehavior, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "mock", "generic":
		return mockOEM{}, nil
	case "supermicro":
		return supermicroOEM{}, nil
	case "dell":
		return dellOEM{}, nil
	case "cisco":
		return ciscoOEM{}, nil
	default:
		return nil, fmt.Errorf("unsupported oem %q (supported: mock, supermicro, dell, cisco)", name)
	}
}

func activeOEM() oemBehavior {
	behavior, err := oemBehaviorFor(config.OEM)
	if err != nil {
		panic(err)
	}
	return behavior
}
