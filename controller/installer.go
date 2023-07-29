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

	"github.com/anik-ghosh-au7/go-pack-node/utils"
)

type Dist struct {
	Tarball string `json:"tarball"`
}

type PackageVersionInfo struct {
	Version string `json:"version"`
	Dist    Dist   `json:"dist"`
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

func Install(args ...string) {
	baseDir, _ := os.Getwd() // Get the current working directory

	if len(args) == 0 {
		// If no args are provided, install all dependencies from dependencies.json
		// ... (implementation omitted for brevity) ...
	} else {
		for _, arg := range args {
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
				continue
			}

			// Check if the version exists in the cache
			cacheDir := filepath.Join(baseDir, ".cache", packageName, packageInfo.Version)
			if !utils.DirExists(cacheDir) {
				// If not, download it and save it in the cache
				err := DownloadPackage(packageInfo, cacheDir)
				if err != nil {
					fmt.Println("Error downloading package:", err)
					continue
				}
			}

			// Update dependencies.json and dependencies-lock.json
			// ... (implementation omitted for brevity) ...
		}
	}
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
