package utils

import (
	"log"
	"os"
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
