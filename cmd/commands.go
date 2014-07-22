package cmd

import (
	"github.com/gonuts/commander"
)

// Commands - Command list.
type Commands []*Command

// Commander - Returns list of *gonuts/commander.Command's.
func (commands Commands) Commander() []*commander.Command {
	list := []*commander.Command{}
	for _, command := range commands {
		list = append(list, command.Commander())
	}
	return list
}
