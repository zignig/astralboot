package packet

import (
	"bytes"
	"testing"
)

func TestAckSerialization(t *testing.T) {
	ack := NewAck(5)
	expack, err := ParsePacket(ack.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if expack.GetType() != ACK {
		t.Fatal("Wrong type!")
	}
	ackpkt, ok := expack.(*AckPacket)
	if !ok {
		t.Fatal("type assertion failed")
	}

	if *ackpkt != 5 {
		t.Fatal("Wrong blocknum")
	}
}

func TestDataSerialization(t *testing.T) {
	dpkt := DataPacket{
		Data:     []byte("hello world"),
		BlockNum: 17,
	}

	expdata, err := ParsePacket(dpkt.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if expdata.GetType() != DATA {
		t.Fatal("Wrong type!")
	}
	datapkt, ok := expdata.(*DataPacket)
	if !ok {
		t.Fatal("type assertion failed")
	}

	if datapkt.BlockNum != dpkt.BlockNum {
		t.Fatal("Wrong blocknum")
	}

	if !bytes.Equal(datapkt.Data, dpkt.Data) {
		t.Fatal("Data mismatch!")
	}
}
