package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/anik-ghosh-au7/go-pack-node/controller"
	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

func main() {
	allowedCommands := []string{"init", "install"}

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

	cacheDir := filepath.Join(dir, ".cache")
	depFile := filepath.Join(dir, "dependencies.json")
	lockFile := filepath.Join(dir, "dependencies-lock.json")
	depDir := filepath.Join(dir, "dependencies")

	switch command {
	case "init":
		var yFlag bool
		if len(os.Args) > 3 && os.Args[len(os.Args)-1] == "-y" {
			yFlag = true
		}
		controller.Initialize(yFlag, cacheDir, depFile, lockFile, depDir, dir)
	case "install":
		var packages []string
		if len(os.Args) > 2 {
			packages = os.Args[2:]
		}
		controller.Install("dependencies", packages...)
	default:
		log.Fatalf("Error: Invalid command")
	}
}
