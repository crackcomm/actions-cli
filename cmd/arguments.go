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

// PushName - Returns `arg.Push` or `arg.Name` if empty.
func (arg *Argument) PushName() string {
	if arg.Push != "" {
		return arg.Push
	}
	return arg.Name
}

// String - Returns arg name in braces eq. "{arg-name}".
func (arg *Argument) String() string {
	return fmt.Sprintf("{%s}", arg.Name)
}

// GoString - Returns GoString used by GoStringer (%#v).
func (arg *Argument) GoString() string {
	return fmt.Sprintf("&%#v", *arg)
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
