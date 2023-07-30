package controller

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
)

func Run(scriptName string) {
	wg := &sync.WaitGroup{}
	// Read the dependencies.json file
	file, err := os.ReadFile("dependencies.json")
	if err != nil {
		log.Fatalf("Error reading dependencies.json: %v", err)
	}

	// Unmarshal the JSON into a Dependency struct
	var deps schema.Dependency
	err = json.Unmarshal(file, &deps)
	if err != nil {
		log.Fatalf("Error parsing dependencies.json: %v", err)
	}

	// Get the script command from the scripts map
	scriptCmd, ok := deps.Scripts[scriptName]
	if !ok {
		log.Fatalf("Error: No such script found to run")
	}

	wg.Add(1)
	// Run the script in a goroutine
	go func() {
		defer wg.Done()
		// Split the command and its arguments
		cmdParts := strings.Split(scriptCmd, " ")
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)

		// Set the command's output to the standard output and standard error
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Run the command
		err := cmd.Run()
		if err != nil {
			log.Fatalf("Error running script: %v", err)
		}
	}()
	wg.Wait()
}
