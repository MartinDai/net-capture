package message

import (
	"github.com/google/gopacket"
)

type NetMessage struct {
	Packet gopacket.Packet
}

func (nm *NetMessage) String() string {
	return nm.Packet.Dump()
}
