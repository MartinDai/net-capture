package input

import (
	"context"
	"errors"
	"net-capture/pkg/listener"
	"net-capture/pkg/logger"
	"net-capture/pkg/message"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrorStopped = errors.New("reading stopped")

type IPInput struct {
	sync.Mutex
	cancelListener context.CancelFunc
	closed         bool
	Expire         time.Duration
	Stats          bool
	Host           string
	Port           uint16
	quit           chan bool
	listener       *listener.IPListener
}

func NewIPInput(address string) (i *IPInput) {
	i = new(IPInput)
	i.Init(address)
	i.listen()
	return
}

func (i *IPInput) Init(address string) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		logger.Fatal(nil, "error while parsing address: %s", address)
	}

	portNum, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Fatal(nil, "error while parsing address: %s", address)
	}

	i.Host = parts[0]
	i.Port = uint16(portNum)

	i.quit = make(chan bool)
}

func (i *IPInput) listen() {
	i.Expire = time.Second * 2
	var err error
	i.listener, err = listener.NewIPListener(i.Host, i.Port, i.Expire)
	if err != nil {
		logger.Fatal(err, "create listener failed")
	}

	err = i.listener.Activate()
	if err != nil {
		logger.Fatal(err, "listen failed")
	}
	var ctx context.Context
	ctx, i.cancelListener = context.WithCancel(context.Background())
	errCh := i.listener.ListenBackground(ctx)
	<-i.listener.Reading
	go func() {
		<-errCh // the listener closed voluntarily
		_ = i.Close()
	}()
}

func (i *IPInput) PluginRead() (*message.NetMessage, error) {
	var msg *message.NetMessage
	select {
	case <-i.quit:
		return nil, ErrorStopped
	case msg = <-i.listener.Messages():
	}

	//这里可以对抓包数据做转换，然后输出自己想要的对象格式
	return msg, nil
}

func (i *IPInput) Close() error {
	i.Lock()
	defer i.Unlock()
	if i.closed {
		return nil
	}
	i.cancelListener()
	close(i.quit)
	i.closed = true
	return nil
}
