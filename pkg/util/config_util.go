package util

import (
	"fmt"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"net"
	"net-capture/pkg/model"
	"path"
	"path/filepath"
	"strings"
)

func GetConfig(configFile string) (*model.Config, error) {
	var config model.Config
	var err error
	if err := loadConfig(configFile, &config); err != nil {
		return nil, err
	}

	if err = checkInput(config.Input); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadConfig(configFile string, config interface{}) error {
	if configFile == "" {
		return fmt.Errorf("config file not specified")
	}

	fileExt := path.Ext(configFile)
	if fileExt != ".yml" && fileExt != ".yaml" {
		return fmt.Errorf("config file only supports .yml or .yaml format")
	}

	absolutePath, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}

	k := koanf.New("::")
	err = k.Load(file.Provider(absolutePath), yaml.Parser())
	if err != nil {
		return err
	}

	err = k.Unmarshal("", config)
	if err != nil {
		return err
	}

	return nil
}

func checkInput(input []model.InputConfig) error {
	if len(input) == 0 {
		return fmt.Errorf("input cannot be empty")
	}

	for _, i := range input {
		if i.Address == "" {
			return fmt.Errorf("input address cannot be empty")
		}

		parts := strings.Split(i.Address, ":")
		if len(parts) > 2 {
			return fmt.Errorf("input address is not valid")
		}

		host := parts[0]
		port := ""

		if len(parts) == 2 {
			port = parts[1]
		}

		if host != "" {
			if host != "localhost" && !isIP(host) {
				return fmt.Errorf("input address host not valid")
			}
		}

		if port == "" {
			return fmt.Errorf("input address must contains port")
		}
	}

	return nil
}

func isIP(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	return true
}
