// Example of minimal DHCP server:
package main

import (
	"fmt"

	dhcp "github.com/krolaw/dhcp4"

	"net"
	"time"
)

// Example using DHCP with a single network interface device
func dhcpServer(c *Config, l *Store) {
	handler := &DHCPHandler{
		ip:            c.BaseIP,
		config:        c,
		leaseDuration: 2 * time.Hour,
		start:         net.IP{192, 168, 2, 2},
		leaseRange:    50,
		leases:        l,
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte{255, 255, 255, 0},
			dhcp.OptionBootFileName:     []byte("undionly.kpxe"),
			dhcp.OptionRouter:           []byte(c.Gateway.To4()),
			dhcp.OptionDomainNameServer: []byte(c.DNSServer.To4()),
		},
	}
	dhcp.ListenAndServeIf(c.Interf, handler)
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
		dhcp.OptionRouter:           []byte(h.config.Gateway.To4()),   // Presuming Server is also your router
		dhcp.OptionDomainNameServer: []byte(h.config.DNSServer.To4()), // Presuming Server is also your DNS server
	}
	IP, err := h.leases.GetIP(p.CHAddr())
	logger.Critical("IP for the lease is ", IP)
	if err != nil {
		fmt.Println("lease get fail ", err)
		return nil
	}
	switch msgType {
	case dhcp.Discover:
		logger.Debug("Discover %s", p.CHAddr())
		return dhcp.ReplyPacket(p, dhcp.Offer, h.config.BaseIP.To4(), IP, h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
	case dhcp.Request:
		logger.Debug("Request %s", p.CHAddr())
		userClass := string(options[77])
		switch userClass {
		// initial hardware boot
		case "iPXE":
			logger.Info("iPXE request")
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			rp.SetSIAddr(h.ip)
			return rp
		// scondary iPXE boot from tftp server
		case "skinny":
			logger.Info("skinny request")
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				skinnyOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			return rp
		default:
			logger.Info("normal dhcp request")
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
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
