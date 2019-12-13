package dao

import "github.com/p14yground/cook/model"

var (
	// Servers ..
	Servers map[string][]*model.Server
	// Tags ..
	Tags []string
	// Config ..
	Config *model.Config
)

func init() {
	Servers = make(map[string][]*model.Server)
}

// LoadConfig ..
func LoadConfig(configFilePath string) error {
	var err error
	Config, err = model.ReadInConfig(configFilePath)
	if err != nil {
		return err
	}
	for i := 0; i < len(Config.Servers); i++ {
		server := Config.Servers[i]
		for j := 0; j < len(Config.Servers[i].Tags); j++ {
			tag := Config.Servers[i].Tags[j]
			if len(Servers[tag]) == 0 {
				Tags = append(Tags, tag)
			}
			Servers[tag] = append(Servers[tag], server)
		}
	}
	return err
}
