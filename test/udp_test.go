package test

import (
	"fmt"
	"net"
	"net-capture/pkg/logger"
	"sync"
	"testing"
)

func TestInputUdp(t *testing.T) {
	wg := new(sync.WaitGroup)

	port := 7777
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
	if err != nil {
		t.Error(err)
		return
	}

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	go func() {
		bytes := make([]byte, 1024)
		for {
			read, err := conn.Read(bytes)
			if err != nil {
				return
			}

			if read == 0 {
				continue
			}

			logger.Info("received: %s", string(bytes[:read]))
			wg.Done()
		}
	}()

	for i := 0; i < 1; i++ {
		wg.Add(1)
		sendDataToLocalUdp(port)
	}

	wg.Wait()
}

func sendDataToLocalUdp(port int) {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "127.0.0.1", port))
	if err != nil {
		logger.Error(err, "Error resolving UDP address")
		return
	}

	// 创建UDP连接
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.Error(err, "Error creating UDP connection")
		return
	}
	defer conn.Close()

	// 向服务器发送数据
	message := []byte("hello")
	// 将数据发送到UDP服务器
	_, err = conn.Write(message)
	if err != nil {
		logger.Error(err, "Error sending data to UDP server")
		return
	}

	logger.Info("Sent: %s", message)
}
