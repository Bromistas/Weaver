package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

func CheckIps() (net.IP, error) {
	// Get the default network interface
	iface, err := net.InterfaceByName("eth0")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Get the IP addresses associated with this interface
	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var ip net.IP
	for _, addr := range addrs {

		if strings.Contains(addr.String(), "::") {
			continue
		}

		switch v := addr.(type) {
		case *net.IPNet:
			//if strings.Contains(v.String(), "::"){
			ip = v.IP
			//}
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil {
			fmt.Printf("%s has IP %s\n", iface.Name, ip.String())
		}
	}

	return ip, nil
}

func TakeBytesAndAdd1(b []byte) []byte {
	hexa := hex.EncodeToString(b)
	parseInt, err := strconv.ParseInt(hexa, 16, 64)

	if err != nil {
		log.Fatalf("Failed to parse int: %v", err)
	}

	parseInt = parseInt + 1
	hexa = strconv.FormatInt(parseInt, 16)

	return []byte(hexa)
}
