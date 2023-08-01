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

var binMutex = &sync.Mutex{}
var depsMutex = &sync.Mutex{}
var fileMutex = &sync.Mutex{}
var copyMutex = &sync.Mutex{}
var installedPackagesMutex = &sync.Mutex{}
var installedPackages = make(map[string]bool)
var wg = &sync.WaitGroup{}

func Install(isRoot bool, args ...string) error {
	baseDir, _ := os.Getwd()
	cacheDir := filepath.Join(baseDir, ".cache")
	depDir := filepath.Join(baseDir, "node_modules")
	depFile := filepath.Join(baseDir, "package.json")
	lockFile := filepath.Join(baseDir, "package-lock.json")

	if _, err := os.Stat(depFile); os.IsNotExist(err) {
		return fmt.Errorf("no package.json file found in project root. please initialize the project first")
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

				installedPackagesMutex.Lock()
				if _, isInstalled := installedPackages[packageName]; isInstalled {
					installedPackagesMutex.Unlock()
					return
				}
				installedPackages[packageName] = true
				installedPackagesMutex.Unlock()

				packageInfo, err := FetchPackageInfo(packageName, packageVersion)
				if err != nil {
					log.Printf("error fetching package info: %v", err)
					return
				}

				cacheDir := filepath.Join(baseDir, ".cache", packageName, packageInfo.Version)
				if !utils.DirExists(cacheDir) {
					err := DownloadPackage(packageInfo, cacheDir, baseDir)
					if err != nil {
						log.Printf("error downloading package: %v", err)
						return
					}
				}

				depsMutex.Lock()
				if isRoot {
					deps.Dependencies[packageName] = packageInfo.Version
				}
				lockDeps.Dependencies[packageName] = &schema.Dependency{
					Version:       packageInfo.Version,
					Resolved:      packageInfo.Dist.Tarball,
					ParentPackage: packageName,
					Dependencies:  packageInfo.Dependencies,
				}
				depsMutex.Unlock()

				depPackageDir := filepath.Join(depDir, packageName)
				if utils.DirExists(depPackageDir) {
					os.RemoveAll(depPackageDir)
				}

				log.Printf("Copying package %s@%s to node_modules...\n", packageName, packageInfo.Version)
				wg.Add(1)
				go func() {
					defer wg.Done()
					copyMutex.Lock()
					err := utils.CopyDir(cacheDir, depPackageDir)
					copyMutex.Unlock()
					if err != nil {
						log.Printf("error copying package: %s", err)
					}
				}()

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

	fileMutex.Lock()
	defer fileMutex.Unlock()
	utils.WriteDepFiles(depFile, lockFile, deps, lockDeps)

	packageLock := schema.PackageLock{
		Dependencies: lockDeps.Dependencies,
	}
	packageLockFile, err := os.Create(filepath.Join(baseDir, "node_modules", ".package-lock.json"))
	if err != nil {
		return err
	}
	defer packageLockFile.Close()

	if err := json.NewEncoder(packageLockFile).Encode(&packageLock); err != nil {
		return err
	}

	return nil
}

func FetchPackageInfo(packageName string, version string) (*schema.PackageVersionInfo, error) {
	encodedPackageName := url.PathEscape(packageName)

	resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", encodedPackageName))
	if err != nil || resp.StatusCode != 200 {
		resp, err = http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/%s", encodedPackageName, version))
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

	if version == "latest" || version == "" {
		version = packageInfo.DistTags.Latest
	}

	versionInfo, exists := packageInfo.Versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found for package %s", version, packageName)
	}

	pkgVersionInfo := &schema.PackageVersionInfo{
		Version:      version,
		Dist:         versionInfo.Dist,
		Dependencies: versionInfo.Dependencies,
	}

	return pkgVersionInfo, nil
}

func DownloadPackage(packageInfo *schema.PackageVersionInfo, destination string, baseDir string) error {
	resp, err := http.Get(packageInfo.Dist.Tarball)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download package: %s", resp.Status)
	}

	err = os.MkdirAll(destination, os.ModePerm)
	if err != nil {
		return err
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error reading tarball: %v", err)
			return err
		}

		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(filepath.Join(destination, strings.TrimPrefix(hdr.Name, "package/")), 0755); err != nil {
				return err
			}
		} else {
			dir := filepath.Join(destination, filepath.Dir(strings.TrimPrefix(hdr.Name, "package/")))
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}

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

	// Binary linking
	packageJsonPath := filepath.Join(destination, "package.json")
	packageJsonFile, err := os.Open(packageJsonPath)
	if err != nil {
		return err
	}
	defer packageJsonFile.Close()

	var packageJson struct {
		Bin map[string]string `json:"bin"`
	}
	if err := json.NewDecoder(packageJsonFile).Decode(&packageJson); err != nil {
		return err
	}

	binDir := filepath.Join(baseDir, "node_modules", ".bin")
	if err := os.MkdirAll(binDir, os.ModePerm); err != nil {
		return err
	}

	for binName, binPath := range packageJson.Bin {
		binLinkPath := filepath.Join(binDir, binName)
		binTargetPath := filepath.Join(destination, binPath)

		binMutex.Lock()
		if err := os.Symlink(binTargetPath, binLinkPath); err != nil {
			binMutex.Unlock()
			return err
		}
		binMutex.Unlock()
	}

	return nil
}
