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
func dhcpServer(c *Config, l *Store) {
	serverIP := c.BaseIP
	handler := &DHCPHandler{
		ip:            serverIP,
		config:        c,
		leaseDuration: 2 * time.Hour,
		start:         net.IP{192, 168, 2, 2},
		leaseRange:    50,
		leases:        l,
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte{255, 255, 255, 0},
			dhcp.OptionBootFileName:     []byte("undionly.kpxe"),
			dhcp.OptionRouter:           []byte(serverIP), // Presuming Server is also your router
			dhcp.OptionDomainNameServer: []byte(serverIP), // Presuming Server is also your DNS server
		},
	}
	fmt.Println("start dhcp")
	log.Fatal(dhcp.ListenAndServeIf(c.Interf, handler)) // Select interface on multi interface device
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
	config        *Config
}

func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	//fmt.Println(p)
	//fmt.Println(p.CHAddr())
	if h.leases.CheckLease(p.CHAddr()) == false {
		h.leases.NewLease(p.CHAddr())
	}
	skinnyOptions := dhcp.Options{
		dhcp.OptionSubnetMask:       []byte{255, 255, 255, 0},
		dhcp.OptionBootFileName:     []byte("http://" + h.ip.String() + "/choose"),
		dhcp.OptionRouter:           []byte(h.ip), // Presuming Server is also your router
		dhcp.OptionDomainNameServer: []byte(h.ip), // Presuming Server is also your DNS server
	}
	IP, err := h.leases.GetIP(p.CHAddr())
	fmt.Println("IP for the lease is ", IP)
	if err != nil {
		fmt.Println("lease get fail ", err)
		return nil
	}
	switch msgType {
	case dhcp.Discover:
		fmt.Println("Discover")
		return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, IP, h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		return nil
	case dhcp.Request:
		fmt.Println("Request")
		userClass := string(options[77])
		switch userClass {
		case "iPXE":
			fmt.Println("iPXE request")
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.ip, net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			rp.SetSIAddr(h.ip)
			return rp
		case "skinny":
			fmt.Println("skinny request")
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.ip, net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				skinnyOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			return rp
		default:
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.ip, net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				skinnyOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			return rp
		}
	case dhcp.Release:
		fmt.Println("Release")
		break
	case dhcp.Decline:
		fmt.Println("Decline")
		break
	}

	return nil
}
