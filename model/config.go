package model

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config ..
type Config struct {
	Servers []Server
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

	viper.OnConfigChange(func(in fsnotify.Event) {
		viper.Unmarshal(&c)
	})

	go viper.WatchConfig()
	return &c, nil
}
