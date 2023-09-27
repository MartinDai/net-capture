package emitter

import (
	"errors"
	"fmt"
	"io"
	"net-capture/pkg/logger"
	"net-capture/pkg/message"
	"net-capture/pkg/plugin"
	"sync"
	"time"
)

func NewEmitter() *Emitter {
	return &Emitter{}
}

type Emitter struct {
	sync.WaitGroup
	plugins *plugin.InOutPlugins
}

// Start initialize loop for sending data from inputs to outputs
func (e *Emitter) Start(plugins *plugin.InOutPlugins) {

	e.plugins = plugins

	for _, in := range plugins.Inputs {
		e.Add(1)
		go func(in message.PluginReader) {
			defer e.Done()
			if err := CopyMulti(in, plugins.Outputs...); err != nil {
				logger.Debug("[EMITTER] error during copy: %q", err)
			}
		}(in)
	}
}

func (e *Emitter) Close() {
	for _, p := range e.plugins.All {
		if cp, ok := p.(io.Closer); ok {
			_ = cp.Close()
		}
	}
	if len(e.plugins.All) > 0 {
		// wait for everything to stop
		e.Wait()
	}
	e.plugins.All = nil // avoid Close to make changes again
}

// CopyMulti copies from 1 reader to multiple writers
func CopyMulti(src message.PluginReader, writers ...message.PluginWriter) (err error) {
	filteredCount := 0
	filteredRequestsLastCleanTime := time.Now().UnixNano()
	filteredRequests := make(map[string]int64)
	for {
		msg, er := src.PluginRead()
		if er != nil {
			if fmt.Sprintf("%v", er.Error()) == "reading stopped" {
				break
			}
			logger.Error(er, "read plugin data error，%v")
			continue
		}

		if msg != nil {
			for _, dst := range writers {
				if err := dst.PluginWrite(msg); err != nil && !errors.Is(err, io.ErrClosedPipe) {
					return err
				}
			}
		}
		// Run GC on each 1000 request
		if filteredCount > 0 && filteredCount%1000 == 0 {
			// Clean up filtered requests for which we didn't get a response to filter
			now := time.Now().UnixNano()
			if now-filteredRequestsLastCleanTime > int64(60*time.Second) {
				for k, v := range filteredRequests {
					if now-v > int64(60*time.Second) {
						delete(filteredRequests, k)
						filteredCount--
					}
				}
				filteredRequestsLastCleanTime = time.Now().UnixNano()
			}
		}
	}

	return err
}
