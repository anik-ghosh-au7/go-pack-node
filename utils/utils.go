package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
)

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	return strings.ToLower(matchAllCap.ReplaceAllString(str, "${1}_${2}"))
}

func CheckOrCreateDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		CheckError(err)
	}
}

func CheckOrCreateFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err = os.Create(path)
		CheckError(err)
	}
}

func CheckError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If the destination directory exists, remove it
	if err == nil {
		os.RemoveAll(dst)
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src string, dst string) error {
	// Check if source file exists
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", src)
	} else if err != nil {
		return fmt.Errorf("error accessing source file: %s", err)
	}

	// Try to open source file
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer in.Close()

	// Try to create destination file
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer out.Close()

	// Try to copy source to destination
	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	// Try to close the destination file
	err = out.Close()
	if err != nil {
		return fmt.Errorf("error closing destination file: %s", err)
	}

	return nil
}

func ReadDepFiles(depFile string, lockFile string) (*schema.Dependency, *schema.Dependency, error) {
	// Initialize empty Dependency and DepLock objects
	dep := &schema.Dependency{}
	lock := &schema.Dependency{}

	// Read the dependencies.json file
	file, err := os.ReadFile(depFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading dependencies file: %s", err)
	}
	if len(file) == 0 {
		return nil, nil, fmt.Errorf("dependencies file is empty")
	}
	err = json.Unmarshal(file, dep)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling dependencies: %s", err)
	}

	// Initialize Dependencies map if it is nil
	if dep.Dependencies == nil {
		dep.Dependencies = make(map[string]string)
	}

	// Read the dependencies-lock.json file
	file, err = os.ReadFile(lockFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading lock file: %s", err)
	}
	if len(file) == 0 {
		return nil, nil, fmt.Errorf("lock file is empty")
	}
	err = json.Unmarshal(file, lock)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling lock: %s", err)
	}

	// Initialize Dependencies map if it is nil
	if lock.Dependencies == nil {
		lock.Dependencies = make(map[string]string)
	}

	return dep, lock, nil
}

// WriteDepFiles writes the given Dependency and DepLock objects to
// dependencies.json and dependencies-lock.json respectively.
func WriteDepFiles(depFile string, lockFile string, dep *schema.Dependency, lock *schema.Dependency) {
	// Marshal dependencies.json
	depData, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(depFile, depData, 0644)
	if err != nil {
		fmt.Println(err)
	}

	// Marshal dependencies-lock.json
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(lockFile, lockData, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func DirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
