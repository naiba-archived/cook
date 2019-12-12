package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/p14yground/cook/model"
)

const welcome = `________________ _______ ______ __
__  ____/__  __ \__  __ \___  //_/
_  /     _  / / /_  / / /__  ,<   
/ /___   / /_/ / / /_/ / _  /| |  
\____/   \____/  \____/  /_/ |_|

Welcome to cook!`

const (
	_ = iota
	cTypeSetTag
)

var servers map[string][]*model.Server
var tags []string
var configFilePath string

func init() {
	servers = make(map[string][]*model.Server)
}

func log(format string, args ...interface{}) {
	fmt.Printf("🐮[LOG]"+format+"\n", args...)
}

func loadConfig() {
	conf, err := model.ReadInConfig(configFilePath)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(conf.Servers); i++ {
		server := conf.Servers[i]
		for j := 0; j < len(conf.Servers[i].Tags); j++ {
			tag := conf.Servers[i].Tags[j]
			if len(servers[tag]) == 0 {
				tags = append(tags, tag)
			}
			servers[tag] = append(servers[tag], &server)
		}
	}
}

var suggests = []prompt.Suggest{
	{Text: "dial", Description: "连接到主机"},
	{Text: "quit", Description: "退出 Cook"},
}

func completer(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
}

func execCommand(command string) {
	args := strings.Split(command, " ")
	switch args[0] {
	case "quit":
		os.Exit(0)
	}
	log("args %v", args)
}

func main() {
	flag.StringVar(&configFilePath, "conf", "./config.yaml", "要加载的配置文件")
	flag.Parse()
	loadConfig()
	fmt.Println(welcome)
	log("load %d servers, %d tags.", len(servers), len(tags))
	log("configFile: %s", configFilePath)
	for {
		t := prompt.Input("", completer, prompt.OptionPrefix("🐮[CMD]> "))
		execCommand(t)
	}
}
