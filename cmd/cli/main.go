package main

import (
	"log"

	"github.com/p14yground/cook/model"
)

func main() {
	conf, err := model.ReadInConfig("./config.yaml")
	if err != nil {
		panic(err)
	}
	log.Printf("conf:%v", conf)
}
