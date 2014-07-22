package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/crackcomm/go-actions/action"
	"github.com/crackcomm/go-actions/local"
	"github.com/crackcomm/go-clitable"
	"github.com/gonuts/commander"
	"io/ioutil"
	"strings"
)

// UsageDescriptionTemplate - Used to construct usage description. (name + args)
var UsageDescriptionTemplate = "%s %s"

// LongDescriptionTemplate - Used to construct long command description. (description + example)
var LongDescriptionTemplate = `%s

example:

    $ %s`

// Command - Command structure.
type Command struct {
	Name        string      `json:"name"`
	Usage       string      `json:"usage"`
	Example     string      `json:"example"`     // one line example
	Description string      `json:"description"`
	Arguments   Arguments   `json:"arguments"`
	IAction     interface{} `json:"action"`      // action to run (string or map)
	Commands    Commands    `json:"commands"`    // list of subcommands
}

// Commander - Creates and returns `commander.Command` structure.
func (cmd *Command) Commander() *commander.Command {
	c := &commander.Command{
		UsageLine: cmd.UsageDescription(),
		Short:     cmd.Description,
		Long:      cmd.LongDescription(),
	}
	// Add handler if action is not empty
	if cmd.IAction != nil {
		c.Run = cmd.Handler
	}
	// Add subcommands if any
	if cmd.Commands != nil {
		c.Subcommands = cmd.Commands.Commander()
	}
	c.Flag.String("format", "table", "result display format")
	c.Flag.Bool("q", false, "only print error and warning messages, all other output will be suppressed")
	return c
}

// Handler - Returns command handler that runs cmd.Action.
func (cmd *Command) Handler(ctx *commander.Command, args []string) (err error) {
	// Parse arguments
	context, err := cmd.ParseArgs(args)
	if err != nil {
		return
	}

	// Run action
	res, err := cmd.RunAction(context)
	if err != nil {
		return
	}

	quiet := ctx.Flag.Lookup("q").Value.Get().(bool)
	if quiet {
		return
	}

	format := ctx.Flag.Lookup("format").Value.Get().(string)

	// If format is json print json and exit
	if format == "json" {
		var body []byte
		body, err = res.PrettyJSON()
		fmt.Printf("%s\n", body)
		return
	}

	// If value is too long - shorten it for display in table
	for key, val := range res {
		var value string
		// Convert value to a string first
		if v, ok := res.Get(key).String(); ok {
			value = v
		} else if v, ok := res.Get(key).Bytes(); ok {
			value = string(v)
		} else if v, ok := res.Get(key).Map(); ok {
			body, _ := v.JSON()
			value = string(body)
		} else {
			value = fmt.Sprintf("%v", val)
		}

		if len(value) >= 100 {
			value = value[:100] + "..."
		}
		res.Add(key, value)
	}

	clitable.PrintHorizontal(res)

	return
}

// ParseArgs - Parses command line arguments into action Context.
func (cmd *Command) ParseArgs(args []string) (res action.Map, err error) {
	res = action.Map{}
	for n, arg := range cmd.Arguments {
		// If no more arguments - return
		if len(args) < n+1 {
			// If argument is required - return error
			if arg.Required {
				err = fmt.Errorf("Argument %s is required.", arg.Name)
			}
			return
		}
		// Get argument value
		var value string
		// If argument is last - get the rest, otherwise just first part
		if len(cmd.Arguments) <= n+1 {
			value = strings.Join(args[n:], " ")
		} else {
			value = args[n]
		}
		// Push value to result
		res.Push(arg.PushName(), value)
	}
	return
}

// Action - Creates and returns action.
func (cmd *Command) Action(ctx action.Map) *action.Action {
	value := action.Format{cmd.IAction}
	// If value is a string it's action name
	if name, ok := value.String(); ok {
		return &action.Action{
			Name: name,
			Ctx:  ctx,
		}
	}
	// If value is a map it's action
	if m, ok := value.Map(); ok {
		// Get action name
		name, _ := m.Get("name").String()
		// Get action ctx
		if c, ok := m.Get("ctx").Map(); ok {
			c.Merge(ctx)
			ctx = c
		}
		return &action.Action{
			Name: name,
			Ctx:  ctx,
		}
	}
	return nil
}

// RunAction - Runs action.
func (cmd *Command) RunAction(ctx action.Map) (action.Map, error) {
	// Get command action
	a := cmd.Action(ctx)
	// Run it
	return local.Run(a)
}

// Run - Runs command.
func (cmd *Command) Run(args []string) error {
	// Get commander command
	app := cmd.Commander()
	// Run it
	return app.Dispatch(args[1:])
}

// UsageDescription - Creates a usage description.
func (cmd *Command) UsageDescription() string {
	// If usage was filled return it
	if cmd.Usage != "" {
		return cmd.Usage
	}
	// If no arguments were set just return name
	if cmd.Arguments == nil {
		return cmd.Name
	}
	// Connect name and arguments into usage line
	return fmt.Sprintf(UsageDescriptionTemplate, cmd.Name, cmd.Arguments.String())
}

// LongDescription - Creates a long description with example.
func (cmd *Command) LongDescription() string {
	example := cmd.Example
	if example == "" {
		example = cmd.Usage
	}
	return fmt.Sprintf(LongDescriptionTemplate, cmd.Description, example)
}

// ReadFile - Reads file unmarshals JSON body and returns a Command.
func ReadFile(filename string) (cmd *Command, err error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	cmd = &Command{}
	err = json.Unmarshal(body, cmd)
	return
}
