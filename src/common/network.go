package common

import (
	"errors"
	"fmt"
	"net"
)

func GetHostIPV1() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", fmt.Errorf("failed to get addresses for interface %v: %v", iface.Name, err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback addresses
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Return the first non-loopback IPv4 address found
			ip = ip.To4()
			if ip != nil {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no IP address found")
}

func GetHostIPV2() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // Interface is down or loopback
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue // Skip interfaces with errors
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Check if the IP is an IPv4 address and not in the Docker network range
			if ip != nil && ip.To4() != nil && !ip.IsLoopback() && !isDockerNetwork(ip) {
				return ip.String(), nil
			}
		}
	}

	return "", errors.New("no suitable IP address found")
}

// isDockerNetwork checks if the IP address belongs to the default Docker network range.
// This function might need adjustments based on your Docker network configuration.
func isDockerNetwork(ip net.IP) bool {
	_, dockerNet, _ := net.ParseCIDR("172.17.0.0/16")
	return dockerNet.Contains(ip)
}
