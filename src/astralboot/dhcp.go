// DHCP server
package main

import (
	dhcp "github.com/krolaw/dhcp4"

	"net"
	"strings"
	"time"
)

func dhcpServer(c *Config, l *Store) {
	handler := &DHCPHandler{
		ip:            c.BaseIP,
		config:        c,
		leaseDuration: 24 * time.Hour,
		leaseRange:    50,
		leases:        l,
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte(c.Subnet.To4()),
			dhcp.OptionBootFileName:     []byte("undionly.kpxe"),
			dhcp.OptionRouter:           []byte(c.Gateway.To4()),
			dhcp.OptionDomainNameServer: []byte(c.DNSServer.To4()),
			dhcp.OptionDomainName:       []byte(c.Domain),
		},
		skinnyOptions: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte(c.Subnet.To4()),
			dhcp.OptionBootFileName:     []byte("http://" + c.BaseIP.String() + "/choose"),
			dhcp.OptionRouter:           []byte(c.Gateway.To4()),
			dhcp.OptionDomainNameServer: []byte(c.DNSServer.To4()),
			dhcp.OptionDomainName:       []byte(c.Domain),
		},
	}
	logger.Error("%v", dhcp.ListenAndServeIf(c.Interf, handler))
}

func SearchDomains(dom string) []byte {
	tmp := strings.Split(dom, ".")
	out := make([]byte, 0)
	for _, i := range tmp {
		out = append(out, byte(len(i)))
		out = append(out, []byte(i)...)
	}
	out = append(out, []byte{0}...)
	logger.Debug("sd : %v", out)
	logger.Debug("sd : %s", tmp)
	return out
}

//DHCPHandler : data structure for the dhcp server
type DHCPHandler struct {
	ip            net.IP        // Server IP to use
	options       dhcp.Options  // Options to send to new DHCP Clients
	skinnyOptions dhcp.Options  // Options to send to skinny DHCP Clients
	start         net.IP        // Start of IP range to distribute
	leaseRange    int           // Number of IPs to distribute (starting from start)
	leaseDuration time.Duration // Lease period
	leases        *Store        // Map to keep track of leases
	config        *Config
}

//ServeDHCP : function for every dhcp request
func (h *DHCPHandler) ServeDHCP(p dhcp.Packet, msgType dhcp.MessageType, options dhcp.Options) (d dhcp.Packet) {
	// get an existing lease or make a new one
	TheLease, err := h.leases.GetLease(p.CHAddr())
	logger.Info("%s has an ip of %s ", TheLease.MAC, TheLease.IP)
	if err != nil {
		logger.Critical("lease get fail , %s", err)
		return nil
	}
	// handle the DHCP transactions
	switch msgType {
	case dhcp.Discover:
		logger.Debug("Discover %s", p.CHAddr())
		return dhcp.ReplyPacket(p, dhcp.Offer, h.config.BaseIP.To4(), TheLease.GetIP(), h.leaseDuration,
			h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
	case dhcp.Request:
		logger.Debug("Request %s", p.CHAddr())
		userClass := string(options[77])
		switch userClass {
		// initial hardware boot
		case "iPXE":
			logger.Notice("iPXE from %s on %v", TheLease.MAC, TheLease.Name)
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			rp.SetSIAddr(h.ip)
			return rp
		// scondary iPXE boot from tftp server
		case "skinny":
			logger.Notice("Booting Machine %s into %s", TheLease.Name, TheLease.Class)
			if TheLease.Active == true {
				h.skinnyOptions[dhcp.OptionHostName] = []byte(TheLease.Name)
				h.skinnyOptions[dhcp.OptionBootFileName] = []byte("http://" + h.ip.String() + "/boot/" + TheLease.Distro + "/${net0/mac}")
			}
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				h.skinnyOptions.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			return rp
		default:
			logger.Info("normal dhcp request")
			if TheLease.Active == true {
				h.skinnyOptions[dhcp.OptionHostName] = []byte(TheLease.Name)
			}
			rp := dhcp.ReplyPacket(p, dhcp.ACK, h.config.BaseIP.To4(), net.IP(options[dhcp.OptionRequestedIPAddress]), h.leaseDuration,
				h.options.SelectOrderOrAll(options[dhcp.OptionParameterRequestList]))
			return rp
		}
	case dhcp.Release:
		logger.Debug("Release")
		break
	case dhcp.Decline:
		logger.Debug("Decline")
		break
	}
	return nil
}
