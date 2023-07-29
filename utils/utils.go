package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func ToSnakeCase(str string) string {
	var result string
	var words []string
	var lastPos int
	str = strings.Trim(str, " ") // Trim spaces

	for i, char := range str {
		if i > 0 && unicode.IsUpper(char) {
			words = append(words, str[lastPos:i])
			lastPos = i
		} else if char == '-' {
			words = append(words, str[lastPos:i])
			lastPos = i + 1
		}
	}

	// Append the last word.
	if lastPos < len(str) {
		words = append(words, str[lastPos:])
	}

	for i, word := range words {
		if i > 0 {
			result += "_"
		}
		result += strings.ToLower(word)
	}

	return result
}

func CheckOrCreateDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		CheckError(err)
	} else if err == nil {
		err = os.RemoveAll(path)
		CheckError(err)
		err = os.Mkdir(path, 0755)
		CheckError(err)
	} else {
		CheckError(err)
	}
}

func CheckOrCreateFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err = os.Create(path)
		CheckError(err)
	} else if err == nil {
		err = os.Remove(path)
		CheckError(err)
		_, err = os.Create(path)
		CheckError(err)
	} else {
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

func DirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
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
	if err == nil {
		return fmt.Errorf("destination already exists")
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
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
