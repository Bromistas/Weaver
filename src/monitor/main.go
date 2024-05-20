package main

import "net"

func main() {
	ring := &Ring{}
	ring.JoinNetwork(net.ParseIP("192.168.1.1"))
	ring.JoinNetwork(net.ParseIP("192.168.1.2"))
	ring.JoinNetwork(net.ParseIP("192.168.1.3"))
}
