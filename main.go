package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

type Dependency struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func main() {
	var yFlag bool
	if len(os.Args) > 1 && os.Args[len(os.Args)-1] == "-y" {
		yFlag = true
		os.Args = os.Args[:len(os.Args)-1]
	}

	if len(os.Args) < 2 {
		log.Fatalf("Error: No path provided")
	}

	dir := os.Args[1]
	cacheDir := filepath.Join(dir, ".cache")
	depFile := filepath.Join(dir, "dependencies.json")
	lockFile := filepath.Join(dir, "dependencies-lock.json")
	depDir := filepath.Join(dir, "dependencies")

	utils.CheckOrCreateDir(cacheDir)
	utils.CheckOrCreateFile(depFile)
	utils.CheckOrCreateFile(lockFile)
	utils.CheckOrCreateDir(depDir)

	defaultName := utils.ToSnakeCase(filepath.Base(dir))

	dep := Dependency{
		Name:            defaultName,
		Version:         "1.0.0",
		Description:     "My App",
		Main:            "index.js",
		Scripts:         map[string]string{"start": "node index.js"},
		Dependencies:    map[string]string{},
		DevDependencies: map[string]string{},
	}

	if !yFlag {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the name of your project: ")
		dep.Name, _ = reader.ReadString('\n')
		fmt.Print("Enter the version of your project: ")
		dep.Version, _ = reader.ReadString('\n')
		fmt.Print("Enter a short description of your project: ")
		dep.Description, _ = reader.ReadString('\n')
		fmt.Print("Enter the entry point to your project: ")
		dep.Main, _ = reader.ReadString('\n')
	}

	data, err := json.MarshalIndent(dep, "", "  ")
	utils.CheckError(err)
	err = os.WriteFile(depFile, data, 0644)
	utils.CheckError(err)

	// Update dependencies-lock.json
	err = os.WriteFile(lockFile, data, 0644)
	utils.CheckError(err)
}
