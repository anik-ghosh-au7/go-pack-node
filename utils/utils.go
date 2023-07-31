package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
)

var fileMutex = &sync.Mutex{}
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	return strings.ToLower(matchAllCap.ReplaceAllString(str, "${1}_${2}"))
}

func CheckOrCreateDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		return err
	}
	return nil
}

func CheckOrCreateFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err = os.Create(path)
		return err
	}
	return nil
}

func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func CopyDir(src string, dst string, wg *sync.WaitGroup) error {
	defer wg.Done() // Make sure to mark this routine as done when it finishes

	// Clean and check the source directory
	src = filepath.Clean(src)
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// Check the destination directory
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Clean up the destination directory if it already exists
	if err == nil {
		os.RemoveAll(dst)
	}

	// Create the destination directory
	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	// Read entries in the source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Create source and destination paths
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Check if the entry is a directory
		if entry.IsDir() {
			// Increment the WaitGroup counter
			wg.Add(1)
			// Recursively copy the sub-directory
			go CopyDir(srcPath, dstPath, wg)
		} else {
			// Increment the WaitGroup counter
			wg.Add(1)
			// Copy the file
			go CopyFile(srcPath, dstPath, wg)
		}
	}
	return nil
}

func CopyFile(src string, dst string, wg *sync.WaitGroup) error {
	defer wg.Done() // Mark this routine as done when it finishes

	// Check and open the source file
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", src)
	} else if err != nil {
		return fmt.Errorf("error accessing source file: %s", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %s", err)
	}
	defer in.Close()

	// Check and create the parent directory of the destination file
	parentDir := filepath.Dir(dst)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating parent directory: %s", err)
		}
	}

	// Create the destination file
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer out.Close()

	// Copy the source file to the destination file
	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	// Close the destination file
	err = out.Close()
	if err != nil {
		return fmt.Errorf("error closing destination file: %s", err)
	}

	return nil
}

func ReadDepFiles(depFile string, lockFile string) (*schema.Dependency, *schema.DependencyLock, error) {
	dep := &schema.Dependency{}
	lock := &schema.DependencyLock{}

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

	if dep.Dependencies == nil {
		dep.Dependencies = make(map[string]string)
	}

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

	if lock.Dependencies == nil {
		lock.Dependencies = make(map[string]*schema.LockDependency)
	}

	return dep, lock, nil
}

func WriteDepFiles(depFile string, lockFile string, dep *schema.Dependency, lock *schema.DependencyLock) {
	// Marshal dependencies.json
	depData, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(depFile, depData, 0644)
	if err != nil {
		fmt.Println(err)
	}

	fileMutex.Lock()
	defer fileMutex.Unlock()

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
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func ReadLockFile(lockFile string) (*schema.LockDependency, error) {
	// Read the file
	fileBytes, err := os.ReadFile(lockFile)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON
	lockDeps := &schema.LockDependency{}
	err = json.Unmarshal(fileBytes, lockDeps)
	if err != nil {
		return nil, err
	}

	return lockDeps, nil
}
