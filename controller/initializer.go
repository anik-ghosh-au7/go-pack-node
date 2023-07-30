package controller

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

func Initialize(yFlag bool, dir string) error {
	cacheDir := filepath.Join(dir, ".cache")
	depFile := filepath.Join(dir, "dependencies.json")
	lockFile := filepath.Join(dir, "dependencies-lock.json")
	depDir := filepath.Join(dir, "node_modules")

	utils.CheckOrCreateDir(cacheDir)
	utils.CheckOrCreateFile(depFile)
	utils.CheckOrCreateFile(lockFile)
	utils.CheckOrCreateDir(depDir)

	defaultName := utils.ToSnakeCase(filepath.Base(dir))

	dep := schema.Dependency{
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
		dep.Name = strings.TrimSpace(dep.Name)
		fmt.Print("Enter the version of your project: ")
		dep.Version, _ = reader.ReadString('\n')
		dep.Version = strings.TrimSpace(dep.Version)
		fmt.Print("Enter a short description of your project: ")
		dep.Description, _ = reader.ReadString('\n')
		dep.Description = strings.TrimSpace(dep.Description)
		fmt.Print("Enter the entry point to your project: ")
		dep.Main, _ = reader.ReadString('\n')
		dep.Main = strings.TrimSpace(dep.Main)
	}

	data, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(depFile, data, 0644)
	if err != nil {
		return err
	}

	// Initialize an empty DependencyLock struct
	lock := schema.DependencyLock{
		Dependencies: make(map[string]*schema.LockDependency),
	}

	// Marshal the lock struct to JSON
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the dependencies-lock.json file
	err = os.WriteFile(lockFile, lockData, 0644)
	if err != nil {
		return err
	}

	return nil
}
