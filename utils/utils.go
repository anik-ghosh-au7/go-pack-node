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
var copyFileMutex = &sync.Mutex{}
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

func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		copyFileMutex.Lock()
		defer copyFileMutex.Unlock()

		return CopyFile(path, dstPath)
	})
}

func CopyFile(srcFile string, dstFile string) (err error) {
	in, err := os.Open(srcFile)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dstFile)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}

	err = out.Sync()
	return
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
