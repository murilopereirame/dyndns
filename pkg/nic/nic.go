package nic

import (
	"fmt"
	"net"
	"net/netip"
)

type NIC struct {
	Interface string
}

func (n *NIC) GetIPv4Address() (string, error) {
	iface, err := net.InterfaceByName(n.Interface)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()

	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr.String())
		if ip != nil && ip.To4() != nil {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found for interface %s", n.Interface)
}

func (n *NIC) GetIPv6Address() (string, error) {
	iface, err := net.InterfaceByName(n.Interface)
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()

	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ip, err := netip.ParseAddr(addr.String())
		if err != nil {
			continue
		}

		if ip.Is6() && ip.IsGlobalUnicast() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv6 address found for interface %s", n.Interface)
}
