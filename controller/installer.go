package controller

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/anik-ghosh-au7/go-pack-node/schema"
	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

var depsMutex = &sync.Mutex{}
var fileMutex = &sync.Mutex{}
var wg = &sync.WaitGroup{}

func Install(isRoot bool, args ...string) {
	baseDir, _ := os.Getwd() // Get the current working directory
	cacheDir := filepath.Join(baseDir, ".cache")
	depDir := filepath.Join(baseDir, "node_modules")
	depFile := filepath.Join(baseDir, "dependencies.json")
	lockFile := filepath.Join(baseDir, "dependencies-lock.json")

	// Check if the dependencies.json file exists
	if _, err := os.Stat(depFile); os.IsNotExist(err) {
		log.Fatalf("No dependencies.json file found in project root. Please initialize the project first")
	}

	// If the cache, dependencies directories or the json files don't exist, create them
	for _, dir := range []string{cacheDir, depDir, lockFile} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if strings.HasSuffix(dir, ".json") {
				// If it's a json file, create the file
				_, err := os.Create(dir)
				if err != nil {
					log.Fatalf("Error creating file: %v", err)
				}
			} else {
				// Else create the directory
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					log.Fatalf("Error creating directory: %v", err)
				}
			}
		}
	}

	// Read the existing dependencies.json and dependencies-lock.json files
	deps, lockDeps, err := utils.ReadDepFiles(depFile, lockFile)

	if err != nil {
		log.Fatalf("Error reading dependency files: %v", err)
	}

	if len(args) == 0 {
		// If no args are provided, install all dependencies from dependencies.json
		depsToInstall := make([]string, 0, len(deps.Dependencies))
		for dep, version := range deps.Dependencies {
			depsToInstall = append(depsToInstall, fmt.Sprintf("%s@%s", dep, version))
		}
		Install(true, depsToInstall...)
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
					log.Fatalf("Error fetching package info: %v", err)
				}

				// Check if the version exists in the cache
				cacheDir := filepath.Join(baseDir, ".cache", packageName, packageInfo.Version)
				if !utils.DirExists(cacheDir) {
					// If not, download it and save it in the cache
					err := DownloadPackage(packageInfo, cacheDir)
					if err != nil {
						log.Fatalf("Error downloading package: %v", err)
					}
				}

				// Update dependencies.json and dependencies-lock.json
				depsMutex.Lock()
				if isRoot {
					deps.Dependencies[packageName] = packageInfo.Version
					lockDeps.Dependencies[packageName] = packageInfo.Version
				} else {
					lockDeps.Dependencies[packageName] = packageInfo.Version
				}
				depsMutex.Unlock()

				// Write the updated dependencies back to the files
				fileMutex.Lock()
				utils.WriteDepFiles(depFile, lockFile, deps, lockDeps)
				fileMutex.Unlock()

				depPackageDir := filepath.Join(depDir, packageName)
				if utils.DirExists(depPackageDir) {
					// If it does, delete the folder
					os.RemoveAll(depPackageDir)
				}

				// Copy the package from cache to the dependencies directory
				err = utils.CopyDir(cacheDir, depPackageDir)
				if err != nil {
					fmt.Printf("Error copying package: %s\n", err)
				}

				// Install dependencies
				for dep := range packageInfo.Dependencies {
					wg.Add(1)
					go func(dep string) {
						defer wg.Done()
						Install(false, dep)
					}(dep)
				}
			}(arg)
		}
	}
	wg.Wait()
}

func FetchPackageInfo(packageName string, version string) (*schema.PackageVersionInfo, error) {
	// Properly encode the package name in the URL
	encodedPackageName := url.PathEscape(packageName)

	// First, try to fetch the latest version of the package
	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", encodedPackageName))
	if err != nil || resp.StatusCode != 200 {
		resp, err = http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", encodedPackageName, version))
		// If that fails, try to fetch the specific version of the package
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch package info: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var packageInfo schema.PackageInfo
	err = json.Unmarshal(body, &packageInfo)
	if err != nil {
		return nil, err
	}

	// If the user didn't provide a specific version, use the latest version
	if version == "latest" || version == "" {
		version = packageInfo.DistTags.Latest
	}

	// Extract the version-specific information
	versionInfo, exists := packageInfo.Versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found for package %s", version, packageName)
	}

	// Create a new PackageVersionInfo object and populate it with the necessary information
	pkgVersionInfo := &schema.PackageVersionInfo{
		Version:      version,
		Dist:         versionInfo.Dist,
		Dependencies: versionInfo.Dependencies,
	}

	return pkgVersionInfo, nil
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

	// Create directory
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
