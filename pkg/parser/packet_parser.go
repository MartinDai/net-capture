package parser

import (
	"github.com/google/gopacket"
	"net"
	"net-capture/pkg/message"
)

type MessageParser struct {
	messages chan *message.NetMessage
	packets  chan gopacket.Packet
	close    chan struct{}
	port     uint16
	ips      []net.IP
}

func NewMessageParser(messages chan *message.NetMessage, port uint16, ips []net.IP) (parser *MessageParser) {
	parser = new(MessageParser)

	parser.messages = messages
	parser.packets = make(chan gopacket.Packet, 1000)
	parser.close = make(chan struct{}, 1)
	parser.port = port
	parser.ips = ips

	go parser.wait()

	return parser
}

func (parser *MessageParser) PacketHandler(packet gopacket.Packet) {
	parser.packets <- packet
}

func (parser *MessageParser) wait() {
	for {
		select {
		case packet := <-parser.packets:
			parser.processPacket(packet)
		case <-parser.close:
			return
		}
	}
}

func (parser *MessageParser) processPacket(packet gopacket.Packet) {
	if packet == nil {
		return
	}

	//可以在这里把packet解析并转换其他结构体
	msg := &message.NetMessage{
		Packet: packet,
	}
	parser.messages <- msg
}
