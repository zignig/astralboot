package client

import (
	"errors"
	"fmt"
	pkt "github.com/whyrusleeping/go-tftp/packet"
	"io"
	"net"
	"time"
)

var ErrTimeout = errors.New("timeout")

type TftpClient struct {
	servaddr  *net.UDPAddr
	udpconn   *net.UDPConn
	packets   chan *packetReceipt
	kill      chan struct{}
	Blocksize int
}

func NewTftpClient(addr string) (*TftpClient, error) {
	laddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	uconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	cli := &TftpClient{
		servaddr:  raddr,
		udpconn:   uconn,
		Blocksize: 512,
		packets:   make(chan *packetReceipt),
		kill:      make(chan struct{}),
	}

	go cli.recvLoop()

	return cli, nil
}

type packetReceipt struct {
	Packet pkt.Packet
	Addr   *net.UDPAddr
	Err    error
}

func (cl *TftpClient) recvLoop() {
	buf := make([]byte, 32768*2)
	for {
		pkt, addr, err := cl.recvPacket(buf)
		select {
		case cl.packets <- &packetReceipt{pkt, addr, err}:
		case <-cl.kill:
			return
		}
	}
}

func (cl *TftpClient) Close() {
	cl.udpconn.Close()
	cl.kill <- struct{}{}
}

func (cl *TftpClient) sendPacket(p pkt.Packet, addr *net.UDPAddr) error {
	data := p.Bytes()
	n, err := cl.udpconn.WriteToUDP(p.Bytes(), addr)
	if err != nil {
		fmt.Printf("Write UDP error: %s\n", err)
		fmt.Printf("attempted to write %d bytes to '%s'\n", len(p.Bytes()), addr)
		fmt.Printf("Packet type was: %d\n", p.GetType())
		fmt.Printf("packet bytes: %v\n", p.Bytes())
		return err
	}

	if n != len(data) {
		return errors.New("Failed to send entire packet")
	}

	return nil
}

func (cl *TftpClient) recvPacket(buf []byte) (pkt.Packet, *net.UDPAddr, error) {
	n, addr, err := cl.udpconn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, err
	}

	if n == len(buf) {
		fmt.Println("Warning! Read entire buffer size, possible errors occurred!")
	}
	buf = buf[:n]

	pkt, err := pkt.ParsePacket(buf)
	if err != nil {
		return nil, nil, err
	}

	return pkt, addr, nil
}

func (cl *TftpClient) PutFile(filename string, data io.Reader) (int, error) {
	req := &pkt.ReqPacket{
		Filename:  filename,
		Mode:      "octet",
		Type:      pkt.WRQ,
		BlockSize: cl.Blocksize,
	}

	err := cl.sendPacket(req, cl.servaddr)
	if err != nil {
		return 0, err
	}

	blknum := uint16(0)
	xferred := 0
	quit := false
	buf := make([]byte, cl.Blocksize)
	var lastPacket pkt.Packet = req
	var addr *net.UDPAddr
	for {
		var recv *packetReceipt

		success := false
		for !success {
			select {
			case recv = <-cl.packets:
				success = true
				addr = recv.Addr
			case <-time.After(time.Second * 5):
				fmt.Println("Receive timeout!")
				var err error
				if addr == nil {
					err = cl.sendPacket(lastPacket, cl.servaddr)
				} else {
					err = cl.sendPacket(lastPacket, addr)
				}
				if err != nil {
					return 0, err
				}
			}
		}
		if recv.Err != nil {
			return 0, recv.Err
		}

		switch p := recv.Packet.(type) {
		case *pkt.ErrorPacket:
			fmt.Println("Error packet.")
			return 0, p
		case *pkt.AckPacket:
			if p.GetBlocknum() != blknum {
				fmt.Printf("Wrong blocknumber! (%d != %d)\n", p.GetBlocknum(), blknum)
				continue
			}
			if blknum == 0 && cl.Blocksize != 512 {
				fmt.Println("Didnt get expected OACK.")
			}
		case *pkt.OAckPacket:
			if blknum != 0 {
				return 0, errors.New("Received OACK at unexpected time...")
			}
			if p.Options["blksize"] != fmt.Sprint(cl.Blocksize) {
				fmt.Printf("Blocksize Negotiation failed!\ngot '%s'\n", p.Options["blocksize"])
			}
		default:
			return 0, fmt.Errorf("unexpected packet: %v, %d", p, p.GetType())
		}
		if quit {
			break
		}
		blknum++
		buf = buf[:cl.Blocksize]
		n, err := data.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		buf = buf[:n]
		xferred += n

		datapkt := &pkt.DataPacket{
			BlockNum: blknum,
			Data:     buf,
		}
		lastPacket = datapkt
		err = cl.sendPacket(datapkt, addr)
		if err != nil {
			return 0, err
		}
		if n < cl.Blocksize {
			quit = true
		}
	}

	return xferred, nil
}

func (cl *TftpClient) GetFile(filename string, out io.Writer) (int, error) {
	req := &pkt.ReqPacket{
		Filename:  filename,
		Mode:      "octet",
		Type:      pkt.RRQ,
		BlockSize: cl.Blocksize,
	}

	err := cl.sendPacket(req, cl.servaddr)
	if err != nil {
		return 0, err
	}

	xfersize := 0
	blknum := uint16(1)
	var lastPacket pkt.Packet = req
	var addr *net.UDPAddr
	for {
		var recv *packetReceipt
		success := false
		for !success {
			select {
			case recv = <-cl.packets:
				addr = recv.Addr
				success = true
			case <-time.After(time.Second * 5):
				var err error
				if addr == nil {
					err = cl.sendPacket(lastPacket, cl.servaddr)
				} else {
					err = cl.sendPacket(lastPacket, addr)
				}
				if err != nil {
					return 0, err
				}
			}
		}

		if err != nil {
			return 0, err
		}

		var data []byte
		switch recv.Packet.GetType() {
		case pkt.ERROR:
			return 0, recv.Packet.(*pkt.ErrorPacket)
		case pkt.DATA:
			datapkt := recv.Packet.(*pkt.DataPacket)
			if datapkt.BlockNum != blknum {
				return 0, fmt.Errorf("Got wrong numbered data packet! (%d != %d)", datapkt.BlockNum, blknum)
			}
			data = datapkt.Data

			// If we have an output writer, write the data out
			if out != nil {
				n, err := out.Write(data)
				if err != nil {
					return xfersize, err
				}
				if n != len(data) {
					return xfersize + n, err
				}
			}
		case pkt.OACK:
			fmt.Println("GOT OACK!!!")
			blknum--
			oack := recv.Packet.(*pkt.OAckPacket)
			if oack.Options["blksize"] != fmt.Sprint(cl.Blocksize) {
				return 0, errors.New("failed to negotiate blocksize")
			}
		default:
			fmt.Printf("Got: %d\n", recv.Packet.GetType())
			fmt.Println(recv.Packet.(*pkt.AckPacket).GetBlocknum())
			fmt.Println(blknum)
			return 0, errors.New("Expected DATA packet!")
		}

		ack := pkt.NewAck(blknum)
		err = cl.sendPacket(ack, addr)
		if err != nil {
			return 0, err
		}

		xfersize += len(data)
		if len(data) < cl.Blocksize {
			break
		}
		blknum++
	}
	return xfersize, nil
}
