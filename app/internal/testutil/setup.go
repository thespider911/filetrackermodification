package testutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SetupTestEnvironment - creates a test_tracker folder on the desktop and copies test files into it (just for testing)
func SetupTestEnvironment() error {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	// Create the test_tracker folder on the desktop
	testFolder := filepath.Join(homeDir, "Desktop", "test_tracker")
	err = os.MkdirAll(testFolder, 0755)
	if err != nil {
		return fmt.Errorf("error creating test folder: %w", err)
	}

	// Get the project root directory
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("error finding project root: %w", err)
	}

	// Define the source test_data folder
	testDataFolder := filepath.Join(projectRoot, "test_data")

	// Copy files from test_data to test_tracker
	err = filepath.Walk(testDataFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(testDataFolder, path)
		if err != nil {
			return fmt.Errorf("error computing relative path: %w", err)
		}

		destPath := filepath.Join(testFolder, relPath)

		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return fmt.Errorf("error creating destination directory: %w", err)
		}

		return copyFile(path, destPath)
	})

	if err != nil {
		return fmt.Errorf("error copying test files: %w", err)
	}

	fmt.Println("Test environment set up successfully in:", testFolder)
	return nil
}

// copyFile - coping file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}

// findProjectRoot - attempts to locate the project root directory
func findProjectRoot() (string, error) {
	// Start from the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	// Navigate up the directory tree until we find the test_data folder
	for {
		if _, err := os.Stat(filepath.Join(dir, "test_data")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root containing test_data folder")
		}
		dir = parent
	}
}
