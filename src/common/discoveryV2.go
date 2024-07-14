package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

// NetDiscover discovers services by broadcasting a message and waiting for responses.
// It accepts a port, a role, an election flag, and a returnAll flag as parameters.
func NetDiscover(port string, role string, election bool, returnAll bool) ([]string, error) {
	timeOut := 10000

	num, _ := strconv.Atoi(port)
	broadcastAddr := net.UDPAddr{
		Port: num,
		IP:   net.IPv4bcast,
	}

	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	message := []byte("Are you a chord?")
	conn.WriteTo(message, &broadcastAddr)

	buffer := make([]byte, 1024)
	var discoveredIPs []string

	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Error("Error setting deadline for incoming messages.")
		return nil, err
	}

	for i := 0; i < timeOut; i++ {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			continue
		}

		if string(buffer[:n]) == fmt.Sprintf("I am a %s chord", role) {
			currentIp := strings.Split(addr.String(), ":")[0]
			if !returnAll {
				if election {
					if len(discoveredIPs) == 0 || CompareIPs(net.ParseIP(currentIp), net.ParseIP(discoveredIPs[0])) == -1 {
						discoveredIPs = []string{currentIp} // Keep only the lowest IP if election is true
					}
				} else {
					return []string{currentIp}, nil // Return immediately if not collecting all IPs
				}
			} else {
				discoveredIPs = append(discoveredIPs, currentIp) // Collect all IPs if returnAll is true
			}
		}
	}

	if len(discoveredIPs) > 0 {
		return discoveredIPs, nil // Return all discovered IPs or the lowest IP if election is true
	}

	log.Infof("No chord of role %s found", role)
	return nil, nil // Return nil if no IPs were discovered
}

// ThreadBroadListen listens for broadcast messages on the specified port.
func ThreadBroadListen(port string, role string) {
	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Error("Error to running udp server")
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Error("Error to read the buffer")
			continue
		}

		message := string(buffer[:n])
		log.Infof("Message receive from %s: %s\n", clientAddr, message)

		if message == "Are you a chord?" {
			response := []byte(fmt.Sprintf("I am a %s chord", role))
			conn.WriteTo(response, clientAddr)
		}

	}

}
