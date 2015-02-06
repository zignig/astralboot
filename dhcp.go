// Example of minimal DHCP server:
package main

import (
	"fmt"

	dhcp "github.com/krolaw/dhcp4"

	"log"
	"net"
	"time"
)

// Example using DHCP with a single network interface device
func dhcpServer(l *Store) {
	serverIP := net.IP{10, 251, 10, 228}
	handler := &DHCPHandler{
		ip:            serverIP,
		leaseDuration: 2 * time.Hour,
		start:         net.IP{172, 30, 0, 2},
		leaseRange:    50,
		leases:        l,
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte{255, 255, 240, 0},
			dhcp.OptionRouter:           []byte(serverIP), // Presuming Server is also your router
			dhcp.OptionDomainNameServer: []byte(serverIP), // Presuming Server is also your DNS server
		},
	}
	fmt.Println("start dhcp")
	log.Fatal(dhcp.ListenAndServeIf("eth0", handler)) // Select interface on multi interface device
	fmt.Println("end dhcp")
}

type lease struct {
	nic    string    // Client's CHAddr
	expiry time.Time // When the lease expires
}

type DHCPHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp.Options  // Options to send to DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        *Store        // Map to keep track of leases
}

func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	fmt.Println(p)
	fmt.Println(p.CHAddr())
	if h.leases.CheckLease(p.CHAddr()) == false {
		h.leases.NewLease(p.CHAddr())
	}
	if options[60] != nil {
		fmt.Println(string(options[60]))
	}
	return nil
}
