package nic

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"strings"
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

	slog.Info(fmt.Sprintf("Found %d addr for interface %s", len(addrs), n.Interface), "addrs", addrs)

	for _, addr := range addrs {
		ipParts := strings.Split(addr.String(), "/")
		ip := net.ParseIP(ipParts[0])
		if ip != nil && ip.To4() != nil && !ip.IsPrivate() {
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

	slog.Info(fmt.Sprintf("Found %d addr for interface %s", len(addrs), n.Interface), "addrs", addrs)

	for _, addr := range addrs {
		ipParts := strings.Split(addr.String(), "/")
		ip, err := netip.ParseAddr(ipParts[0])
		if err != nil {
			slog.Warn("Invalid IP received", "ip", addr.String())
			continue
		}

		if ip.Is6() && ip.IsGlobalUnicast() && !ip.IsPrivate() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv6 address found for interface %s", n.Interface)
}
