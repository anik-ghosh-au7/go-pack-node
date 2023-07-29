package controller

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

func Install(args ...string) {
	wg := &sync.WaitGroup{}
	baseDir, _ := os.Getwd() // Get the current working directory
	cacheDir := filepath.Join(baseDir, ".cache")
	depDir := filepath.Join(baseDir, "dependencies")
	depFile := filepath.Join(baseDir, "dependencies.json")
	lockFile := filepath.Join(baseDir, "dependencies-lock.json")

	// If the cache, dependencies directories or the json files don't exist, create them
	for _, dir := range []string{cacheDir, depDir, depFile, lockFile} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if strings.HasSuffix(dir, ".json") {
				// If it's a json file, create the file
				_, err := os.Create(dir)
				if err != nil {
					fmt.Println("Error creating file:", err)
					return
				}
			} else {
				// Else create the directory
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					fmt.Println("Error creating directory:", err)
					return
				}
			}
		}
	}

	// Read the existing dependencies.json and dependencies-lock.json files
	deps, lockDeps := utils.ReadDepFiles(depFile, lockFile)

	if len(args) == 0 {
		// If no args are provided, install all dependencies from dependencies.json
		depsToInstall := make([]string, 0, len(deps.Dependencies))
		for dep, version := range deps.Dependencies {
			depsToInstall = append(depsToInstall, fmt.Sprintf("%s@%s", dep, version))
		}
		Install(depsToInstall...)
		return
	} else {
		for _, arg := range args {
			wg.Add(1)
			go func(arg string) { // Start a goroutine
				defer wg.Done()
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
				cacheDir := filepath.Join(baseDir, ".cache", packageName, packageInfo.Version)
				if !utils.DirExists(cacheDir) {
					// If not, download it and save it in the cache
					err := DownloadPackage(packageInfo, cacheDir)
					if err != nil {
						fmt.Println("Error downloading package:", err)
						return
					}
				}

				// Update dependencies.json and dependencies-lock.json
				deps.Dependencies[packageName] = packageInfo.Version

				// Write the updated dependencies back to the files
				utils.WriteDepFiles(depFile, lockFile, deps, lockDeps)

				depPackageDir := filepath.Join(depDir, packageName)
				if utils.DirExists(depPackageDir) {
					// If it does, delete the folder
					os.RemoveAll(depPackageDir)
				}

				// Copy the package from cache to the dependencies directory
				err = utils.CopyDir(cacheDir, depPackageDir)
				if err != nil {
					fmt.Println("Error copying package:", err)
				}

				// Install dependencies
				for dep := range packageInfo.Dependencies {
					wg.Add(1)
					go func(dep string) {
						defer wg.Done()
						Install(depDir, dep)
					}(dep)
				}
			}(arg)
		}
	}
	wg.Wait()
}

func FetchPackageInfo(packageName string, version string) (*schema.PackageVersionInfo, error) {
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

	var packageInfo schema.PackageVersionInfo
	err = json.Unmarshal(body, &packageInfo)
	if err != nil {
		return nil, err
	}

	return &packageInfo, nil
}

func DownloadPackage(packageInfo *schema.PackageVersionInfo, destination string) error {
	resp, err := http.Get(packageInfo.Dist.Tarball)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download package: %s", resp.Status)
	}

	// Create directory instead of a file
	err = os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		return err
	}

	// Create a new gzip reader
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Create a new tar reader
	tr := tar.NewReader(gzr)

	// Iterate through the files in the tarball
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Create the directories in the path
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filepath.Join(destination, strings.TrimPrefix(hdr.Name, "package/")), 0755); err != nil {
				return err
			}
		} else {
			// Create the directory path
			dir := filepath.Join(destination, filepath.Dir(strings.TrimPrefix(hdr.Name, "package/")))
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}

			// Create the file
			outFile, err := os.Create(filepath.Join(destination, strings.TrimPrefix(hdr.Name, "package/")))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
