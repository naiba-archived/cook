package model

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config ..
type Config struct {
	Servers []*Server
}

// ReadInConfig ..
func ReadInConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var c Config

	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(c.Servers); i++ {
		if c.Servers[i].Port == "" {
			c.Servers[i].Port = "22"
		}
		if c.Servers[i].User == "" {
			c.Servers[i].User = "root"
		}
		if c.Servers[i].Password == "" && c.Servers[i].IdentityFile == "" {
			c.Servers[i].IdentityFile = "~/.ssh/id_rsa"
		}
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		viper.Unmarshal(&c)
	})

	go viper.WatchConfig()
	return &c, nil
}
