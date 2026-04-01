package utils

import (
	"os"
	"strings"
)

func IsMarkdownFile(filename string) bool {
	return strings.HasSuffix(filename, ".md")
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func WriteFile(filename string, content []byte) error {
	return os.WriteFile(filename, content, 0644)
}

func FilterMarkdownFiles(files []string) []string {
	var mdFiles []string
	for _, f := range files {
		if IsMarkdownFile(f) {
			mdFiles = append(mdFiles, f)
		}
	}
	return mdFiles
}
