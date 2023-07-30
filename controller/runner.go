package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
)

func Run(scriptName string) error {
	wg := &sync.WaitGroup{}
	file, err := os.ReadFile("dependencies.json")
	if err != nil {
		return fmt.Errorf("error reading dependencies.json: %v", err)
	}

	var deps schema.Dependency
	err = json.Unmarshal(file, &deps)
	if err != nil {
		return fmt.Errorf("error parsing dependencies.json: %v", err)
	}

	scriptCmd, ok := deps.Scripts[scriptName]
	if !ok {
		return fmt.Errorf("error: no such script found to run")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmdParts := strings.Split(scriptCmd, " ")
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Printf("error running script: %v", err)
		}
	}()
	wg.Wait()

	return nil
}
