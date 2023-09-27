package output

import (
	"net-capture/pkg/logger"
	"net-capture/pkg/message"
)

// StdOutput used for debugging, prints all incoming requests
type StdOutput struct {
}

// NewStdOutput constructor for StdOutput
func NewStdOutput() (i *StdOutput) {
	i = new(StdOutput)
	return
}

func (i *StdOutput) PluginWrite(msg *message.NetMessage) error {
	logger.Info("================================================================")
	logger.Info("StdOutput Write NetMessage\n%s", msg.String())
	logger.Info("================================================================")
	return nil
}

func (i *StdOutput) String() string {
	return "Std Output"
}
