package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

type Dist struct {
	Tarball string `json:"tarball"`
}

type PackageVersionInfo struct {
	Version string `json:"version"`
	Dist    Dist   `json:"dist"`
}

func Install(args ...string) {
	baseDir, _ := os.Getwd() // Get the current working directory
	var wg sync.WaitGroup    // Create a WaitGroup

	if len(args) == 0 {
		// If no args are provided, install all dependencies from dependencies.json
		// ... (implementation omitted for brevity) ...
	} else {
		for _, arg := range args {
			wg.Add(1) // Add a count to the WaitGroup for each argument

			go func(arg string) { // Wrap the logic for each argument in a goroutine
				defer wg.Done() // Decrement the WaitGroup count when the goroutine finishes

				packageAndVersion := strings.Split(arg, "@")
				packageName := packageAndVersion[0]
				packageVersion := "latest"
				if len(packageAndVersion) > 1 {
					packageVersion = packageAndVersion[1]
				}

				// Fetch package info from npm registry
				packageInfo, err := FetchPackageInfo(packageName, packageVersion)
				if err != nil {
					fmt.Println("Error fetching package info:", err)
					return
				}

				// Check if the version exists in the cache
				cacheDir := filepath.Join(baseDir, ".cache", fmt.Sprintf("%s@%s", packageName, packageInfo.Version))
				if !utils.DirExists(cacheDir) {
					// If not, download it and save it in the cache
					err := DownloadPackage(packageInfo, cacheDir)
					if err != nil {
						fmt.Println("Error downloading package:", err)
						return
					}
				}

				// Update dependencies.json and dependencies-lock.json
				// ... (implementation omitted for brevity) ...
			}(arg)
		}
	}

	wg.Wait() // Wait for all goroutines to finish
}

func FetchPackageInfo(packageName string, version string) (*PackageVersionInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", packageName, version))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch package info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var packageInfo PackageVersionInfo
	err = json.Unmarshal(body, &packageInfo)
	if err != nil {
		return nil, err
	}

	return &packageInfo, nil
}

func DownloadPackage(packageInfo *PackageVersionInfo, destination string) error {
	resp, err := http.Get(packageInfo.Dist.Tarball)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download package: %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
