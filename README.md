# actions-cli

Application for constructing command line applications using only .json files.

To construct a command line application we need two things

1. [Actions](https://bitbucket.org/actions/go-actions)
2. Commands

## Example

Example json commands

```JSON
{
  "name":        "music",
  "usage":       "music command {args...}",
  "description": "Looks for music on the internet",
  "commands": [
    {
      "name":        "youtube",
      "action":      "youtube.find",
      "example":     "music youtube 2pac",
      "description": "Looks for music on YouTube",
      "arguments": [
        {"name": "title", "required": true, "push": "query.search_query"}
      ]
    }
  ]
}
```

Result

```
$ music
music - Looks for music on the internet

Commands:

    youtube     Looks for music on YouTube

Use "music help <command>" for more information about a command.


Options:
  -format="table": result display format
  -q=false: only print error and warning messages, all other output will be suppressed
$ music help youtube
Usage: music youtube {title}

Looks for music on YouTube

example:

    $ music youtube 2pac

Options:
  -format="table": result display format
  -q=false: only print error and warning messages, all other output will be suppressed
$ music youtube 2pac letter to my unborn child
|-----------------------------------------------------------------------------------------------------------------------|
| title       | 2pac-Tupac Letter To My UnBorn                                                                          |
|-----------------------------------------------------------------------------------------------------------------------|
| url         | http://www.youtube.com/watch?v=SyBjzRrEdE0                                                              |
|-----------------------------------------------------------------------------------------------------------------------|
| author      | 2Pac for life                                                                                           |
|-----------------------------------------------------------------------------------------------------------------------|
| views       | 4362727                                                                                                 |
|-----------------------------------------------------------------------------------------------------------------------|
| keywords    | 2pac, tupac, letter, to, my, unborn                                                                     |
|-----------------------------------------------------------------------------------------------------------------------|
| description | Album:Until The End Of Time Disk 1 Song"Letter To My UnBorn" By:Tupac A. Shakur Tha Song And Tha Pic    |
|-----------------------------------------------------------------------------------------------------------------------|
| image       | http://i1.ytimg.com/vi/SyBjzRrEdE0/hqdefault.jpg                                                        |
|-----------------------------------------------------------------------------------------------------------------------|
| rating      | {"dislikes":"129","likes":"13609"}                                                                      |
|-----------------------------------------------------------------------------------------------------------------------|
```

## Usage

Application is JSON marshaled set of commands.

Command structure is following

```Go
type Argument struct {
	Name     string `json:"name"`
	Push     string `json:"push"` // Name in context
	Required bool   `json:"required"`
	Value    string `json:"default"`
}

type Command struct {
	Name        string      `json:"name"`
	Usage       string      `json:"usage"`
	Example     string      `json:"example"`     // one line example
	Description string      `json:"description"`
	Arguments   Arguments   `json:"arguments"`
	IAction     interface{} `json:"action"`      // action to run (string or map)
	Commands    Commands    `json:"commands"`    // list of subcommands
}
```

To use a .json application first you have to register a source of actions.

For example a file source in directory `./actions`.

```Go
package main

import (
	"github.com/crackcomm/go-actions/core"
	"github.com/crackcomm/go-actions/source/file"
	// ...
)

// ...

func init() {
	// Create new file source
	fileSource := &file.Source{"./actions"}
	// Add file source to default core registry
	core.Default.Registry.Sources.Add(fileSource)
}
```

Now `youtube/find.json` file becomes `youtube.find` action and you can use it in your application.

Application now can be unmarshaled from a file and runned.

```Go
package main

import (
	// ...
	"github.com/crackcomm/actions-cli/cmd"
	"log"
	"os"
)

// ...

func main() {
	// Read application from .json file
	app, err := cmd.ReadFile("./app.json")
	if err != nil {
		log.Fatal(err)
	}

	// Run application
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
```
