package common

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"net"
	"time"
)

type DiscoveryCallback func(entry *zeroconf.ServiceEntry)

func Discover(serviceType, domain string, timeout time.Duration, callback DiscoveryCallback) error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("Failed to initialize resolver: %v", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			callback(entry)
		}
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = resolver.Browse(ctx, serviceType, domain, entries)
	if err != nil {
		return fmt.Errorf("Failed to browse: %v", err)
	}

	// Wait until the context times out
	<-ctx.Done()
	log.Println("Service discovery finished")

	return nil
}

func RegisterForDiscovery(serviceName, serviceType, domain string, port int, ip net.IP) error {
	iface, err := getInterfaceByIP(ip)
	if err != nil {
		return fmt.Errorf("Failed to get interface: %v", err)
	}

	server, err := zeroconf.Register(
		serviceName,
		serviceType,
		domain,
		port,
		[]string{"txtv=1", "txt2=2"},
		[]net.Interface{*iface},
	)
	if err != nil {
		return fmt.Errorf("Failed to register service: %v", err)
	}
	defer server.Shutdown()

	log.Println("Service registered, waiting to exit...")
	// Keep the service running
	select {}

	return nil
}

func getInterfaceByIP(ip net.IP) (*net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ipNet *net.IPNet
			if ipnet, ok := addr.(*net.IPNet); ok {
				ipNet = ipnet
			}

			if ipNet.IP.Equal(ip) {
				return &i, nil
			}
		}
	}

	return nil, fmt.Errorf("No interface with IP %v found", ip)
}
