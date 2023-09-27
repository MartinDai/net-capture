package model

type Config struct {
	DebugMode bool          `koanf:"debug_mode"`
	Input     []InputConfig `koanf:"input"`
}

type InputConfig struct {
	Address string `koanf:"address"`
}
