package message

type PluginReader interface {
	PluginRead() (message *NetMessage, err error)
}

type PluginWriter interface {
	PluginWrite(message *NetMessage) (err error)
}
