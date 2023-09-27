package test

import (
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestInputIPv4(t *testing.T) {
	wg := new(sync.WaitGroup)

	listener, err := net.Listen("tcp4", "127.0.0.1:6666")
	if err != nil {
		t.Error(err)
		return
	}
	origin := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("this is response"))
			wg.Done()
		}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go origin.Serve(listener)
	defer listener.Close()
	_, port, _ := net.SplitHostPort(listener.Addr().String())

	addr := "http://127.0.0.1:" + port
	for i := 0; i < 1; i++ {
		wg.Add(1)
		_, err = http.Get(addr)

		if err != nil {
			t.Error(err)
			return
		}
	}

	wg.Wait()
}
