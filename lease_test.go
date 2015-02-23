package main

import (
	"fmt"
	"net"
	"testing"
)

func TestLeases(t *testing.T) {
	iplist := NetList(net.IP{192, 168, 6, 0}, net.IP{255, 255, 255, 0})
	fmt.Println(iplist)
}
