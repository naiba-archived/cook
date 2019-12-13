package cmd

import (
	"strings"

	"github.com/c-bata/go-prompt"

	"github.com/p14yground/cook/dao"
)

var commandSuggests = []prompt.Suggest{
	{Text: "connect", Description: "连接到主机"},
	{Text: "quit", Description: "退出 Cook"},
}

var connectSuggests = []prompt.Suggest{
	{Text: "--all", Description: "所有主机"},
	{Text: "--tags", Description: "选择标签"},
}

var connectSuggestsIndex = map[string]struct{}{
	"--all":  {},
	"--tags": {},
}

// NewCompleter ..
func NewCompleter() *Completer {
	c := &Completer{}
	for i := 0; i < len(dao.Tags); i++ {
		c.tagSuggests = append(c.tagSuggests, prompt.Suggest{
			Text: dao.Tags[i],
		})
	}
	return c
}

// Completer ..
type Completer struct {
	tagSuggests []prompt.Suggest
}

// Complete ..
func (c *Completer) Complete(d prompt.Document) []prompt.Suggest {
	command := strings.TrimSpace(d.TextBeforeCursor())
	if command == "" {
		return []prompt.Suggest{}
	}
	args := strings.Split(command, " ")
	switch args[0] {
	case "connect":
		if len(args) == 2 {
			if _, ok := connectSuggestsIndex[args[1]]; !ok {
				return prompt.FilterHasPrefix(connectSuggests, "--", true)
			}
		} else if len(args) == 3 {
			switch args[1] {
			case "--tags":
				tags := strings.Split(args[2], ",")
				if len(tags) > 1 {
					return prompt.FilterHasPrefix(c.tagSuggests, d.GetWordBeforeCursorUntilSeparatorIgnoreNextToCursor(","), true)
				}
				return prompt.FilterHasPrefix(c.tagSuggests, d.GetWordBeforeCursor(), true)
			}
		}
	default:
		return prompt.FilterHasPrefix(commandSuggests, d.GetWordBeforeCursor(), true)
	}
	return []prompt.Suggest{}
}
