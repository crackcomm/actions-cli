package main

import (
	"github.com/crackcomm/actions-cli/cmd"
	_ "github.com/crackcomm/go-core/actions"
	_ "github.com/crackcomm/go-core/filter"
	_ "github.com/crackcomm/go-core/html"
	_ "github.com/crackcomm/go-core/http"
	_ "github.com/crackcomm/go-core/log"
	"log"
	"os"
)

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
