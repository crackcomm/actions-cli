package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/crackcomm/go-actions/action"
	"github.com/crackcomm/go-actions/core"
	"github.com/crackcomm/go-actions/local"
	"github.com/crackcomm/go-actions/source/file"
	"github.com/crackcomm/go-actions/source/http"
	"github.com/crackcomm/go-clitable"
	"github.com/gonuts/commander"
	"gopkg.in/v1/yaml"
	"path/filepath"
	"io/ioutil"
	"net/url"
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
	Name        string      `json:"name" yaml:"name"`
	Usage       string      `json:"usage" yaml:"usage"`
	Example     string      `json:"example" yaml:"example"` // one line example
	Description string      `json:"description" yaml:"description"`
	Arguments   Arguments   `json:"arguments" yaml:"arguments"`
	Flags       Arguments   `json:"flags" yaml:"flags"`
	IAction     interface{} `json:"action" yaml:"action"`   // action to run (string or map)
	Commands    Commands    `json:"commands" yaml:"commands"` // list of subcommands
	Sources     []string    `json:"sources" yaml:"sources"`  // list of actions sources
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

	// Add default flags
	c.Flag.String("format", "table", "result display format")
	c.Flag.Bool("q", false, "only print error and warning messages, all other output will be suppressed")

	// Bind flags
	if cmd.Flags == nil {
		return c
	}

	for _, flag := range cmd.Flags {
		if flag == nil {
			continue
		}
		c.Flag.String(flag.Name, flag.Value, flag.Description)
	}

	return c
}

// Handler - Returns command handler that runs cmd.Action.
func (cmd *Command) Handler(ctx *commander.Command, args []string) (err error) {
	// Parse arguments
	context, err := cmd.ParseContext(ctx, args)
	if err != nil {
		return
	}

	// Bind sources
	cmd.BindSources()

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

// ParseContext - Parses command line arguments into action Context.
func (cmd *Command) ParseContext(ctx *commander.Command, args []string) (res action.Map, err error) {
	res = action.Map{}

	// Extract arguments from `args`
	for n, arg := range cmd.Arguments {
		// If no more arguments - return
		if len(args) < n+1 {
			break
		}

		// Get argument value
		// If argument is last - get the rest, otherwise just first part
		var value string
		if len(cmd.Arguments) <= n+1 {
			value = strings.Join(args[n:], " ")
		} else {
			value = args[n]
		}

		// Set default value if empty
		if value == "" {
			value = arg.Value
		}

		// Push value to result
		res.Push(arg.PushName(), value)
	}

	// Extract flags from `ctx`
	for _, flag := range cmd.Flags {
		// Get value from flag
		value := ctx.Flag.Lookup(flag.Name).Value.Get()

		// Pass if it's empty string or nil
		if v, ok := value.(string); ok && v == "" || value == nil {
			continue
		}

		// Push value to result
		res.Push(flag.PushName(), value)
	}

	// Check for required arguments
	for _, arg := range cmd.Arguments {
		// If argument is empty but required - return error
		if res.Pull(arg.PushName()).IsNil() && arg.Required {
			err = fmt.Errorf("Argument %s is required.", arg.Name)
			return
		}
	}

	// Check for required flags
	for _, flag := range cmd.Flags {
		// If flag is empty but required - return error
		if res.Pull(flag.PushName()).IsNil() && flag.Required {
			err = fmt.Errorf("Flag %s is required.", flag.Name)
			return
		}
	}

	return
}

// Action - Creates and returns action.
func (cmd *Command) Action(ctx action.Map) *action.Action {
	value := action.Format{Value: cmd.IAction}
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

// BindSources - Binds command actions sources.
func (cmd *Command) BindSources() {
	for _, source := range cmd.Sources {
		// If source is a valid url - create http source
		if isURL(source) {
			core.AddSource(&http.Source{Path: source})
		} else {
			// Add file source to default core registry
			core.AddSource(&file.Source{Path: source})
		}
	}
}

// Run - Runs command.
func (cmd *Command) Run(args []string) error {
	// Get commander command
	app := cmd.Commander()

	// Bind sources
	cmd.BindSources()

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

// GoString - Returns GoString used by GoStringer (%#v).
func (cmd *Command) GoString() string {
	return fmt.Sprintf("&%#v", *cmd)
}

// ReadFile - Reads file unmarshals JSON or YAML body and returns a Command.
func ReadFile(filename string) (cmd *Command, err error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	cmd = &Command{}

	switch filepath.Ext(filename) {
	case ".json":
		err = json.Unmarshal(body, cmd)
	case ".yaml":
		err = yaml.Unmarshal(body, cmd)
	}

	return
}

// isURL - Returns true if value url scheme is a `http` or `https`.
func isURL(value string) (yes bool) {
	if uri, err := url.Parse(value); err == nil {
		if uri.Scheme == "http" || uri.Scheme == "https" {
			yes = true
		}
	}
	return
}
