package main

import (
	"log"
	"os"

	"github.com/anik-ghosh-au7/go-pack-node/controller"
	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

func main() {
	allowedCommands := []string{"init", "install", "start", "run"}

	if len(os.Args) < 2 {
		log.Fatalf("Error: Not enough arguments provided")
	}

	command := os.Args[1]
	if !utils.Contains(allowedCommands, command) {
		log.Fatalf("Error: Command not found")
	}

	dir := "."
	if len(os.Args) > 2 {
		dir = os.Args[2]
	}
	if dir == "." {
		dir, _ = os.Getwd()
	}

	var err error
	switch command {
	case "init":
		var yFlag bool
		if len(os.Args) > 3 && os.Args[len(os.Args)-1] == "-y" {
			yFlag = true
		}
		err = controller.Initialize(yFlag, dir)
	case "install":
		var packages []string
		if len(os.Args) > 2 {
			packages = os.Args[2:]
		}
		err = controller.Install(true, packages...)
	case "start":
		if len(os.Args) > 2 {
			log.Fatalf("Error: Invalid 'start' command")
		}
		script := os.Args[1]
		err = controller.Run(script)
	case "run":
		if len(os.Args) < 3 {
			log.Fatalf("Error: Not enough arguments provided for 'run' command")
		}
		script := os.Args[2]
		err = controller.Run(script)
	default:
		log.Fatalf("Error: Invalid command")
	}

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
