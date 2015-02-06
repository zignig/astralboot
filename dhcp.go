// Example of minimal DHCP server:
package main

import (
	"fmt"

	dhcp "github.com/krolaw/dhcp4"

	"log"
	"net"
	"time"
)

func info() {
	in, _ := net.Interfaces()
	for i := range in {
		fmt.Println(in[i])
		fmt.Println(in[i].Addrs())
	}
}

// Example using DHCP with a single network interface device
func dhcpServer(c *Config, l *Store) {
	serverIP := net.IP{192, 168, 2, 1}
	handler := &DHCPHandler{
		ip:            serverIP,
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
	log.Fatal(dhcp.ListenAndServeIf("eth1", handler)) // Select interface on multi interface device
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
	switch msgType {
	case dhcp.Discover:
		fmt.Println("Discover")
		if options[60] != nil {
			vendor := string(options[60])
			if vendor == "PXEClient:Arch:00000:UNDI:002001" {
				fmt.Println("OFFER")
				return dhcp.ReplyPacket(p, dhcp.Offer, h.ip, net.IP{192, 168, 2, 2}, h.leaseDuration,
					h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			}
		}
		return nil
	case dhcp.Request:
		fmt.Println("Request")
		//t := net.IP(h.ip).To4()
		//n := copy(p[20:24], t)
		rp := dhcp.ReplyPacket(p, dhcp.ACK, h.ip, net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
		rp.SetSIAddr(h.ip)
		return rp
	case dhcp.Release:
		fmt.Println("Release")
		break
	case dhcp.Decline:
		fmt.Println("Decline")
		break
	}

	return nil
}

func cidr() {
	ip, ipnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		log.Fatal(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		fmt.Println(ip)
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
