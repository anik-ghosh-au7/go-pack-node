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

func Install(isRoot bool, args ...string) error {
	baseDir, _ := os.Getwd()
	cacheDir := filepath.Join(baseDir, ".cache")
	depDir := filepath.Join(baseDir, "node_modules")
	depFile := filepath.Join(baseDir, "dependencies.json")
	lockFile := filepath.Join(baseDir, "dependencies-lock.json")

	if _, err := os.Stat(depFile); os.IsNotExist(err) {
		return fmt.Errorf("no dependencies.json file found in project root. please initialize the project first")
	}

	for _, dir := range []string{cacheDir, depDir, lockFile} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if strings.HasSuffix(dir, ".json") {
				_, err := os.Create(dir)
				if err != nil {
					return fmt.Errorf("error creating file: %v", err)
				}
			} else {
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return fmt.Errorf("error creating directory: %v", err)
				}
			}
		}
	}

	deps, lockDeps, err := utils.ReadDepFiles(depFile, lockFile)

	if err != nil {
		return fmt.Errorf("error reading dependency files: %v", err)
	}

	if len(args) == 0 {
		depsToInstall := make([]string, 0, len(deps.Dependencies))
		for dep, version := range deps.Dependencies {
			depsToInstall = append(depsToInstall, fmt.Sprintf("%s@%s", dep, version))
		}
		return Install(true, depsToInstall...)
	} else {
		for _, arg := range args {
			wg.Add(1)
			go func(arg string) {
				defer wg.Done()
				packageAndVersion := strings.Split(arg, "@")
				packageName := packageAndVersion[0]
				packageVersion := "latest"
				if len(packageAndVersion) > 1 {
					packageVersion = packageAndVersion[1]
				}

				packageInfo, err := FetchPackageInfo(packageName, packageVersion)
				if err != nil {
					log.Printf("error fetching package info: %v", err)
					return
				}

				cacheDir := filepath.Join(baseDir, ".cache", packageName, packageInfo.Version)
				if !utils.DirExists(cacheDir) {
					err := DownloadPackage(packageInfo, cacheDir)
					if err != nil {
						log.Printf("error downloading package: %v", err)
						return
					}
				}

				depsMutex.Lock()
				if isRoot {
					deps.Dependencies[packageName] = packageInfo.Version
					lockDeps.Dependencies[packageName] = packageInfo.Version
				} else {
					lockDeps.Dependencies[packageName] = packageInfo.Version
				}
				depsMutex.Unlock()

				fileMutex.Lock()
				utils.WriteDepFiles(depFile, lockFile, deps, lockDeps)
				fileMutex.Unlock()

				depPackageDir := filepath.Join(depDir, packageName)
				if utils.DirExists(depPackageDir) {
					os.RemoveAll(depPackageDir)
				}

				err = utils.CopyDir(cacheDir, depPackageDir)
				if err != nil {
					log.Printf("error copying package: %s", err)
					return
				}

				for dep := range packageInfo.Dependencies {
					wg.Add(1)
					go func(dep string) {
						defer wg.Done()
						_ = Install(false, dep)
					}(dep)
				}
			}(arg)
		}
	}
	wg.Wait()

	return nil
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
