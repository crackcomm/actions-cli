package cmd

import (
	"fmt"
	"strings"
)

// Argument - Argument structure.
type Argument struct {
	Name        string `json:"name"`
	Push        string `json:"push"` // Name in context
	Required    bool   `json:"required"`
	Value       string `json:"default"`
	Description string `json:"description"`
}

// Arguments - Map of arguments.
type Arguments []*Argument

// PushName - Returns `a.Push` or `a.Name` if empty.
func (a *Argument) PushName() string {
	if a.Push != "" {
		return a.Push
	}
	return a.Name
}

// String - Returns arg name in braces eq. "{arg-name}".
func (a *Argument) String() string {
	return fmt.Sprintf("{%s}", a.Name)
}

// Get - Gets argument by name.
func (args Arguments) Get(name string) *Argument {
	for _, arg := range args {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}

// String - Joins argument names into a list eq. "{arg1} {arg2}".
func (args Arguments) String() string {
	l := []string{}
	for _, arg := range args {
		l = append(l, arg.String())
	}
	return strings.Join(l, " ")
}
