package dao

import "github.com/p14yground/cook/model"

var (
	// Servers ..
	Servers map[string][]*model.Server
	// Tags ..
	Tags []string
)

func init() {
	Servers = make(map[string][]*model.Server)
}

// LoadConfig ..
func LoadConfig(configFilePath string) error {
	conf, err := model.ReadInConfig(configFilePath)
	if err != nil {
		return err
	}
	for i := 0; i < len(conf.Servers); i++ {
		server := conf.Servers[i]
		for j := 0; j < len(conf.Servers[i].Tags); j++ {
			tag := conf.Servers[i].Tags[j]
			if len(Servers[tag]) == 0 {
				Tags = append(Tags, tag)
			}
			Servers[tag] = append(Servers[tag], &server)
		}
	}
	return nil
}
