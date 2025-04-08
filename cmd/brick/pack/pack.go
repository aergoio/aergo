package pack

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Set to track imported files
var importedFiles = make(map[string]bool)

// ProcessLines processes a string containing contract code, handling imports
func processLines(input string, currentDir string) (string, error) {

	// Create a string builder for the output
	var output strings.Builder

	// Process each line
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		// Check if line starts with "import "
		if strings.HasPrefix(line, "import ") {
			// Extract the file name from import statement
			importLine := strings.TrimSpace(line[6:]) // Remove "import " prefix

			// Check if it has minimum length for a valid import
			if len(importLine) < 3 {
				return "", fmt.Errorf("invalid import format: %s", line)
			}

			// Check if it starts with a valid quote character
			quoteChar := importLine[0]
			if quoteChar != '"' && quoteChar != '\'' {
				return "", fmt.Errorf("import statement must use quotes: %s", line)
			}

			// Check if it ends with the same quote character
			if importLine[len(importLine)-1] != quoteChar {
				return "", fmt.Errorf("mismatched quotes in import: %s", line)
			}

			// Extract the file path between quotes
			importFile := importLine[1 : len(importLine)-1]

			// If it's not a URL and not an absolute path, make it relative to the current file's directory
			if !strings.HasPrefix(importFile, "http") && !filepath.IsAbs(importFile) {
				importFile = filepath.Join(currentDir, importFile)
			}

			// Get absolute path to check for circular imports
			absImportPath, err := filepath.Abs(importFile)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path: %w", err)
			}

			// Skip if already imported
			if importedFiles[absImportPath] {
				continue
			}

			// Mark as imported
			importedFiles[absImportPath] = true

			// Read the imported file
			importContent, err := readContractFile(importFile)
			if err != nil {
				return "", fmt.Errorf("error importing file '%s': %w", importFile, err)
			}

			// Get the directory of the imported file for nested imports
			importedFileDir := filepath.Dir(importFile)

			// Process the imported content recursively
			processedImport, err := processLines(importContent, importedFileDir)
			if err != nil {
				return "", err
			}

			// Add the processed import to output
			output.WriteString(processedImport)
			output.WriteString("\n")
		} else {
			// Regular line, add to output
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning input: %w", err)
	}

	return output.String(), nil
}

func readContractFile(filePath string) (string, error) {
	// if the file path is a url, read it from the web
	if strings.HasPrefix(filePath, "http") {
		// search in the web
		req, err := http.NewRequest("GET", filePath, nil)
		if err != nil {
			return "", err
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		fileBytes, _ := ioutil.ReadAll(resp.Body)
		return string(fileBytes), nil
	}

	// search in the local file system
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(fileBytes), nil
}

func ReadContract(filePath string) (string, error) {
	// Reset imported files tracking for each new processing
	importedFiles = make(map[string]bool)

	// Get the absolute path of the main file
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	// Get the directory of the main file
	mainFileDir := filepath.Dir(absPath)

	// Mark the main file as imported
	importedFiles[absPath] = true

	// read the contract file
	output, err := readContractFile(filePath)
	if err != nil {
		return "", err
	}

	// process the contract file for import statements, passing the main file's directory
	return processLines(output, mainFileDir)
}
