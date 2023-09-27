package listener

import (
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"io"
	"net"
	"net-capture/pkg/logger"
	"net-capture/pkg/message"
	"net-capture/pkg/parser"
	"runtime"
	"strings"
	"sync"
	"time"
)

type IPListener struct {
	sync.Mutex
	messages        chan *message.NetMessage
	host            string
	port            uint16
	trackResponse   bool
	Interfaces      []pcap.Interface
	Reading         chan bool
	Handles         map[string]packetHandle
	closeDone       chan struct{}
	quit            chan struct{}
	expiry          time.Duration
	allowIncomplete bool
	loopIndex       int
	Activate        func() error
	ReadPcap        func()
}

type packetHandle struct {
	handler      *pcap.Handle
	packetSource *gopacket.PacketSource
	ips          []net.IP
}

func NewIPListener(host string, port uint16, expiry time.Duration) (l *IPListener, err error) {
	l = &IPListener{}
	err = l.Init(host, port, expiry)
	if err != nil {
		return nil, err
	}

	l.messages = make(chan *message.NetMessage, 10000)
	l.ReadPcap = l.readPcap
	return
}

func (l *IPListener) Init(host string, port uint16, expiry time.Duration) (err error) {
	l.host = host
	if l.host == "localhost" {
		l.host = "127.0.0.1"
	}
	l.Handles = make(map[string]packetHandle)
	l.closeDone = make(chan struct{})
	l.quit = make(chan struct{})
	l.Reading = make(chan bool)
	l.port = port
	l.expiry = expiry
	l.Activate = l.activatePcap
	err = l.setInterfaces()
	return
}

func (l *IPListener) activatePcap() error {
	var e error
	var msg string
	for _, ifi := range l.Interfaces {
		var handle *pcap.Handle
		handle, e = l.PcapHandle(ifi)
		if e != nil {
			msg += "\n" + e.Error()
			continue
		}

		source := gopacket.NewPacketSource(handle, handle.LinkType())
		source.Lazy = true
		source.NoCopy = true
		l.Handles[ifi.Name] = packetHandle{
			handler:      handle,
			packetSource: source,
			ips:          interfaceIPs(ifi),
		}
	}
	if len(l.Handles) == 0 {
		return fmt.Errorf("pcap handles error:%s", msg)
	}
	return nil
}

// PcapHandle returns new pcap Handle from dev on success.
// this function should be called after setting all necessary options for this listener
func (l *IPListener) PcapHandle(ifi pcap.Interface) (handle *pcap.Handle, err error) {
	var inactive *pcap.InactiveHandle
	inactive, err = pcap.NewInactiveHandle(ifi.Name)
	if err != nil {
		return nil, fmt.Errorf("inactive handle error: %q, interface: %q", err, ifi.Name)
	}
	defer inactive.CleanUp()

	if err = inactive.SetPromisc(true); err != nil {
		return nil, fmt.Errorf("promiscuous mode error: %q, interface: %q", err, ifi.Name)
	}

	var snap = 64<<10 + 200
	err = inactive.SetSnapLen(snap)
	if err != nil {
		return nil, fmt.Errorf("snapshot length error: %q, interface: %q", err, ifi.Name)
	}
	err = inactive.SetTimeout(2000 * time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("handle buffer timeout error: %q, interface: %q", err, ifi.Name)
	}
	handle, err = inactive.Activate()
	if err != nil {
		return nil, fmt.Errorf("PCAP Activate device error: %q, interface: %q", err, ifi.Name)
	}

	bpfFilter := l.Filter(ifi)
	logger.Info("Interface: %s. BPF Filter: %s", ifi.Name, bpfFilter)
	err = handle.SetBPFFilter(bpfFilter)
	if err != nil {
		handle.Close()
		return nil, fmt.Errorf("BPF filter error: %q%s, interface: %q", err, bpfFilter, ifi.Name)
	}
	return
}

func (l *IPListener) readPcap() {
	l.Lock()
	defer l.Unlock()
	for key, handle := range l.Handles {
		go func(key string, ph packetHandle) {
			runtime.LockOSThread()

			defer l.closeHandles(key)

			messageParser := parser.NewMessageParser(l.messages, l.port, ph.ips)

			for {
				select {
				case <-l.quit:
					return
				default:
					packet, err := ph.packetSource.NextPacket()
					if err == io.EOF {
						return
					} else if err != nil {
						if err != pcap.NextErrorTimeoutExpired {
							logger.Error(err, "NextPacket error:")
						}
						continue
					}

					messageParser.PacketHandler(packet)
				}
			}
		}(key, handle)
	}
	close(l.Reading)
}

func (l *IPListener) Filter(ifi pcap.Interface) (filter string) {
	// 如果需要对host做过滤，可以扩展下面的代码
	//hosts := []string{l.host}
	//if listenAll(l.host) || isDevice(l.host, ifi) {
	//	hosts = interfaceAddresses(ifi)
	//}
	//
	//var hostFilters []string
	//hostFilters = append(hostFilters, hostsFilter("dst", hosts))
	//hostFilters = append(hostFilters, hostsFilter("src", hosts))

	var portFilters []string

	portFilters = append(portFilters, portFilter("tcp", "dst", l.port))
	portFilters = append(portFilters, portFilter("udp", "dst", l.port))
	portFilters = append(portFilters, portFilter("tcp", "src", l.port))
	portFilters = append(portFilters, portFilter("udp", "src", l.port))

	return strings.Join(portFilters, " or ")
}

func (l *IPListener) setInterfaces() (err error) {
	var pifis []pcap.Interface
	pifis, err = pcap.FindAllDevs()
	ifis, _ := net.Interfaces()
	l.Interfaces = []pcap.Interface{}

	if err != nil {
		return
	}

	for _, pi := range pifis {
		if strings.HasPrefix(l.host, "k8s://") {
			if !strings.HasPrefix(pi.Name, "veth") {
				continue
			}
		}

		if isDevice(l.host, pi) {
			l.Interfaces = []pcap.Interface{pi}
			return
		}

		var ni net.Interface
		for _, i := range ifis {
			if i.Name == pi.Name {
				ni = i
				break
			}

			addrs, _ := i.Addrs()
			for _, a := range addrs {
				for _, pa := range pi.Addresses {
					if a.String() == pa.IP.String() {
						ni = i
						break
					}
				}
			}
		}

		if ni.Flags&net.FlagLoopback != 0 {
			l.loopIndex = ni.Index
		}

		if runtime.GOOS != "windows" {
			if len(pi.Addresses) == 0 {
				continue
			}

			if ni.Flags&net.FlagUp == 0 {
				continue
			}
		}

		l.Interfaces = append(l.Interfaces, pi)
	}
	return
}

func (l *IPListener) ListenBackground(ctx context.Context) chan error {
	err := make(chan error, 1)
	go func() {
		defer close(err)
		if e := l.Listen(ctx); err != nil {
			err <- e
		}
	}()
	return err
}

func (l *IPListener) Listen(ctx context.Context) (err error) {
	l.ReadPcap()
	done := ctx.Done()
	select {
	case <-done:
		close(l.quit) // signal close on all handles
		<-l.closeDone // wait all handles to be closed
		err = ctx.Err()
	case <-l.closeDone: // all handles closed voluntarily
	}
	return
}

func (l *IPListener) closeHandles(key string) {
	l.Lock()
	defer l.Unlock()
	if handle, ok := l.Handles[key]; ok {
		handle.handler.Close()

		delete(l.Handles, key)
		if len(l.Handles) == 0 {
			close(l.closeDone)
		}
	}
}

func (l *IPListener) Messages() chan *message.NetMessage {
	return l.messages
}

func interfaceIPs(ifi pcap.Interface) []net.IP {
	var ips []net.IP
	for _, addr := range ifi.Addresses {
		ips = append(ips, addr.IP)
	}
	return ips
}

func listenAll(addr string) bool {
	switch addr {
	case "", "0.0.0.0", "[::]", "::":
		return true
	}
	return false
}

func isDevice(addr string, ifi pcap.Interface) bool {
	// Windows npcap loopback have no IPs
	if addr == "127.0.0.1" && ifi.Name == `\Device\NPF_Loopback` {
		return true
	}

	if addr == ifi.Name {
		return true
	}

	if strings.HasSuffix(addr, "*") {
		if strings.HasPrefix(ifi.Name, addr[:len(addr)-1]) {
			return true
		}
	}

	for _, _addr := range ifi.Addresses {
		if _addr.IP.String() == addr {
			return true
		}
	}

	return false
}

func interfaceAddresses(ifi pcap.Interface) []string {
	var hosts []string
	for _, addr := range ifi.Addresses {
		hosts = append(hosts, addr.IP.String())
	}
	return hosts
}

func portFilter(transport string, direction string, port uint16) string {
	if port == 0 {
		return fmt.Sprintf("(%s %s portrange 0-%d)", transport, direction, 1<<16-1)
	}

	return fmt.Sprintf("(%s %s port %d)", transport, direction, port)
}

func hostsFilter(direction string, hosts []string) string {
	var hostsFilters []string
	for _, host := range hosts {
		hostsFilters = append(hostsFilters, fmt.Sprintf("%s host %s", direction, host))
	}

	return strings.Join(hostsFilters, " or ")
}
