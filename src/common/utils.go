package common

import (
	"net"
	"os"
	"sort"
	"syscall"
)

func insertIntoSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] >= item })
	slice = append(slice, 0)
	copy(slice[i+1:], slice[i:])
	slice[i] = item
	return slice
}

func removeFromSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] > item })
	if i < len(slice) && slice[i] == item {
		slice = append(slice[:i], slice[i+1:]...)
	}
	return slice
}

func searchInSorted(slice []int, item int) int {
	return sort.Search(len(slice), func(i int) bool { return slice[i] > item })
}

func CompareIPs(ip1, ip2 net.IP) int {
	for i := 0; i < len(ip1) && i < len(ip2); i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}

func createReusableUDPConn(port string) (net.PacketConn, error) {
	// Resolve the address
	addr, err := net.ResolveUDPAddr("udp4", ":"+port)
	if err != nil {
		return nil, err
	}

	// Create the UDP socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}

	// Set SO_REUSEADDR to allow multiple applications to listen on the same port
	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("setsockopt", err)
	}

	// Bind the socket to the address
	err = syscall.Bind(fd, &syscall.SockaddrInet4{Port: addr.Port})
	if err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("bind", err)
	}

	// Convert the file descriptor back to a net.PacketConn
	file := os.NewFile(uintptr(fd), "")
	conn, err := net.FilePacketConn(file)
	if err != nil {
		syscall.Close(fd)
		return nil, err
	}
	file.Close() // Close the file to avoid leaking the file descriptor.

	return conn, nil
}
