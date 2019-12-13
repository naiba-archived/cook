package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/p14yground/cook/pkg/log"
)

// Executor ..
type Executor struct {
}

// Exec ..
func (e *Executor) Exec(s string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	} else if s == "quit" || s == "exit" {
		fmt.Println("Bye!")
		os.Exit(0)
		return
	}

	log.Printf("exec:%s", s)

}
