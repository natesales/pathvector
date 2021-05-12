package main

import (
	"fmt"
	"net"
)

// UnmarshalYAML implements the interface from go-yaml to marshal an IP address or prefix in CIDR notation
func (a *addr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw string
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	ip, ipNet, err := net.ParseCIDR(raw)
	if err != nil {
		return err
	}

	netMask, _ := ipNet.Mask.Size()
	*a = addr{
		Address: ip,
		Mask:    uint8(netMask),
	}

	return nil
}

// String converts an Addr to a CIDR string
func (a addr) String() string {
	return fmt.Sprintf("%s/%d", a.Address.String(), a.Mask)
}
