// Package app-build implements static builder for actions commands.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/crackcomm/actions-cli/cmd"

	_ "github.com/crackcomm/go-actions/source/file"
	_ "github.com/crackcomm/go-actions/source/http"
	_ "github.com/crackcomm/go-core/actions"
	_ "github.com/crackcomm/go-core/filter"
	_ "github.com/crackcomm/go-core/html"
	_ "github.com/crackcomm/go-core/http"
	_ "github.com/crackcomm/go-core/log"
)

// AppFile - File containing application.
var AppFile = "app.json"

// OutputFile - Output binary file path.
var OutputFile = "app.json"

var appBase = `package main

import (
	"github.com/crackcomm/actions-cli/cmd"
	_ "github.com/crackcomm/go-actions/source/file"
	_ "github.com/crackcomm/go-actions/source/http"
	_ "github.com/crackcomm/go-core"
	"log"
	"os"
)

var app = %s

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
`

var cmdBase = `&cmd.Command{
	Name:        %#v,
	Usage:       %#v,
	Example:     %#v,
	Description: %#v,
	IAction:     %s,
	Sources:     %#v,
	Flags:       cmd.Arguments{%s},
	Arguments:   cmd.Arguments{%s},
	Commands:    cmd.Commands{%s},
}`

var argBase = `&cmd.Argument{
	Name:        %#v,
	Push:        %#v,
	Value:       %#v,
	Required:    %#v,
	Description: %#v,
}`

func argToCode(arg *cmd.Argument, i int) string {
	code := fmt.Sprintf(argBase, arg.Name, arg.Push, arg.Value, arg.Required, arg.Description)
	return indentCode(code, i)
}

func argsToCode(args cmd.Arguments, i int) string {
	if len(args) == 0 {
		return ""
	}
	res := []string{}
	for _, n := range args {
		res = append(res, argToCode(n, 1)+",\n")
	}
	code := strings.Join(res, "")
	return "\n" + indentCode(code, i)
}

func cmdsToCode(commands cmd.Commands, i int) string {
	if len(commands) == 0 {
		return ""
	}

	cmds := []string{}
	for _, n := range commands {
		cmds = append(cmds, cmdToCode(n, 1)+",\n")
	}
	code := strings.Join(cmds, "")
	return "\n" + indentCode(code, i)
}

func cmdToCode(c *cmd.Command, i int) string {
	var a string
	if c.IAction == nil {
		a = "nil"
	} else {
		a = fmt.Sprintf("%#v", c.IAction)
	}
	code := fmt.Sprintf(cmdBase,
		c.Name,
		c.Usage,
		c.Example,
		c.Description,
		a,
		c.Sources,
		argsToCode(c.Flags, 1),
		argsToCode(c.Arguments, 1),
		cmdsToCode(c.Commands, 1),
	)
	code = indentCode(code, i)
	return code
}

func indentCode(code string, i int) string {
	c := strings.Split(code, "\n")
	ind := strings.Repeat("\t", i)
	res := []string{}
	for _, l := range c {
		l = ind + l
		res = append(res, l)
	}
	return strings.Join(res, "\n")
}

func appGen(c *cmd.Command) string {
	code := cmdToCode(c, 0)
	return fmt.Sprintf(appBase, code)
}

func buildApp(c *cmd.Command, output string) (err error) {
	// Create temporary build directory
	builddir := filepath.Join(os.TempDir(), fmt.Sprintf("app-%s-build", c.Name))
	err = os.MkdirAll(builddir, os.ModePerm)
	if err != nil {
		return
	}

	// Generate app source code
	app := appGen(c)

	// Application source code filename
	appfile := filepath.Join(builddir, "main.go")

	// Write application source code to file
	err = ioutil.WriteFile(appfile, []byte(app), os.ModePerm)

	// Build application
	command := exec.Command("go", "build", "-o", output, appfile)
	err = command.Run()

	// Remove temporary dir
	os.RemoveAll(builddir)

	return
}

func main() {
	flag.StringVar(&OutputFile, "o", "Output file", OutputFile)
	flag.StringVar(&AppFile, "app", "Application file (json)", AppFile)
	flag.Parse()

	// Read application from .json file
	app, err := cmd.ReadFile(AppFile)
	if err != nil {
		log.Fatal(err)
	}

	// Build application
	err = buildApp(app, OutputFile)
	if err != nil {
		log.Fatal(err)
	}
}
