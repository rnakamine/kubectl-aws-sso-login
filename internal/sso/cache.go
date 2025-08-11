package sso

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetSSOCacheDir returns the SSO cache directory path
func GetSSOCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".aws", "sso", "cache"), nil
}

// FindSessionFiles finds all JSON files in the SSO cache directory
func FindSessionFiles() ([]string, error) {
	cacheDir, err := GetSSOCacheDir()
	if err != nil {
		return nil, err
	}

	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("SSO cache directory does not exist: %s", cacheDir)
	}

	pattern := filepath.Join(cacheDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find cache files: %w", err)
	}

	return files, nil
}

// LoadSession loads and parses an SSO session from a file
func LoadSession(filename string) (*SSOSession, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	session, err := ParseSessionFromJSON(data)
	if err != nil {
		return nil, err
	}

	// Only return sessions that have an access token
	// (skip client registration cache files)
	if session.AccessToken == "" {
		return nil, fmt.Errorf("not a valid session file (no access token)")
	}

	return session, nil
}

// FindValidSession searches for a valid SSO session in cache files
func FindValidSession() error {
	files, err := FindSessionFiles()
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no SSO cache files found")
	}

	for _, file := range files {
		session, err := LoadSession(file)
		if err != nil {
			// Skip files that can't be parsed or don't have access tokens
			continue
		}

		if session.IsValid() {
			return nil
		}
	}

	return fmt.Errorf("no valid SSO session found")
}
