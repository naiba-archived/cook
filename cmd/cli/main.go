package main

import (
	"flag"
	"fmt"

	"github.com/c-bata/go-prompt"
	pc "github.com/c-bata/go-prompt/completer"

	"github.com/p14yground/cook/cmd/cli/cmd"
	"github.com/p14yground/cook/dao"
	"github.com/p14yground/cook/pkg/log"
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

var configFilePath string

func cli() {
	executor := &cmd.Executor{}
	completer := cmd.NewCompleter()
	p := prompt.New(
		executor.Exec,
		completer.Complete,
		prompt.OptionTitle("Cook：批量管理 SSH 主机"),
		prompt.OptionPrefix("🐮> "),
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionWordSeparator(pc.FilePathCompletionSeparator),
	)
	p.Run()
}

func main() {
	flag.StringVar(&configFilePath, "conf", "./config.yaml", "要加载的配置文件")
	flag.Parse()
	if err := dao.LoadConfig(configFilePath); err != nil {
		panic(err)
	}
	fmt.Println(welcome)
	log.Printf("load %d servers, %d tags.", len(dao.Config.Servers), len(dao.Tags))
	log.Printf("configFile: %s", configFilePath)

	cli()
}
