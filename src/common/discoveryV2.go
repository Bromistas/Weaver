package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"strings"
	"time"
)

// NetDiscover discovers the service by broadcasting a message and waiting for a response.
// It accepts a port and a timeout in seconds as parameters.
func NetDiscover(port string, role string) (string, error) {
	timeOut := 10000

	num, _ := strconv.Atoi(port)
	broadcastAddr := net.UDPAddr{
		Port: num,
		IP:   net.IPv4bcast,
	}

	conn, err := net.ListenPacket("udp4", fmt.Sprintf(":%s", port))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	message := []byte("Are you a chord?")
	conn.WriteTo(message, &broadcastAddr)

	buffer := make([]byte, 1024)

	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		log.Error("Error setting deadline for incoming messages.")
		return "", err
	}

	for i := 0; i < timeOut; i++ {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			continue
		}

		if string(buffer[:n]) == fmt.Sprintf("I am a %s chord", role) {
			foundIp := strings.Split(addr.String(), ":")[0]
			log.Infof("Discover a chord of role %s in %s", role, foundIp)
			return foundIp, nil
		}
	}

	log.Infof("Not found a chord of role %s", role)

	return "", nil

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
