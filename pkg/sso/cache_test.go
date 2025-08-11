package sso

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestCacheDir(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cacheDir := filepath.Join(tmpDir, ".aws", "sso", "cache")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test cache directory: %v", err)
	}

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return cacheDir, cleanup
}

func TestGetSSOCacheDir(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	dir, err := GetSSOCacheDir()
	if err != nil {
		t.Errorf("GetSSOCacheDir() error = %v", err)
		return
	}

	expected := filepath.Join(tmpDir, ".aws", "sso", "cache")
	if dir != expected {
		t.Errorf("GetSSOCacheDir() = %v, want %v", dir, expected)
	}
}

func TestFindSessionFiles(t *testing.T) {
	cacheDir, cleanup := setupTestCacheDir(t)
	defer cleanup()

	tests := []struct {
		name      string
		files     map[string]string
		wantCount int
		wantErr   bool
	}{
		{
			name: "typical SSO cache directory",
			files: map[string]string{
				"d033e22ae348aeb5660fc2140aec35850c4da997.json": "{}",
				"botocore-client-id-us-east-1.json":             "{}",
				"botocore-client-id-us-west-2.json":             "{}",
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "mixed file types",
			files: map[string]string{
				"d033e22ae348aeb5660fc2140aec35850c4da997.json": "{}",
				"config.txt": "text",
				"data.xml":   "<xml/>",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty directory",
			files:     map[string]string{},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean cache directory
			files, _ := filepath.Glob(filepath.Join(cacheDir, "*"))
			for _, f := range files {
				os.Remove(f)
			}

			for filename, content := range tt.files {
				path := filepath.Join(cacheDir, filename)
				err := os.WriteFile(path, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			foundFiles, err := FindSessionFiles()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindSessionFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(foundFiles) != tt.wantCount {
				t.Errorf("FindSessionFiles() found %d files, want %d", len(foundFiles), tt.wantCount)
			}
		})
	}
}

func TestLoadSession(t *testing.T) {
	cacheDir, cleanup := setupTestCacheDir(t)
	defer cleanup()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid session with access token",
			content: `{
				"startUrl": "https://example.awsapps.com/start",
				"region": "us-east-1",
				"accessToken": "valid-token-123",
				"expiresAt": "2024-12-31T23:59:59Z"
			}`,
			wantErr: false,
		},
		{
			name: "client registration file (no access token)",
			content: `{
				"clientId": "client-123",
				"clientSecret": "secret-456",
				"expiresAt": "2024-12-31T23:59:59Z"
			}`,
			wantErr: true, // LoadSession rejects files without accessToken
		},
		{
			name:    "invalid JSON",
			content: `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(cacheDir, "test-session.json")
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err = LoadSession(filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSession() error = %v, wantErr %v", err, tt.wantErr)
			}

			os.Remove(filename)
		})
	}
}

func TestFindValidSession(t *testing.T) {
	cacheDir, cleanup := setupTestCacheDir(t)
	defer cleanup()

	futureTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	pastTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name    string
		files   map[string]string
		wantErr bool
	}{
		{
			name: "valid session with client registration files",
			files: map[string]string{
				"botocore-client-id-us-east-1.json": `{
					"clientId": "client-123",
					"clientSecret": "secret-456",
					"region": "us-east-1"
				}`,
				"d033e22ae348aeb5660fc2140aec35850c4da997.json": fmt.Sprintf(`{
					"startUrl": "https://example.awsapps.com/start",
					"region": "us-east-1",
					"accessToken": "valid-token",
					"expiresAt": "%s"
				}`, futureTime),
			},
			wantErr: false,
		},
		{
			name: "expired session with client registration files",
			files: map[string]string{
				"botocore-client-id-us-east-1.json": `{
					"clientId": "client-123",
					"clientSecret": "secret-456"
				}`,
				"d033e22ae348aeb5660fc2140aec35850c4da997.json": fmt.Sprintf(`{
					"accessToken": "expired-token",
					"expiresAt": "%s"
				}`, pastTime),
			},
			wantErr: true,
		},
		{
			name: "only client registration files",
			files: map[string]string{
				"botocore-client-id-us-east-1.json": `{
					"clientId": "client-1",
					"clientSecret": "secret-1"
				}`,
				"botocore-client-id-us-west-2.json": `{
					"clientId": "client-2",
					"clientSecret": "secret-2"
				}`,
			},
			wantErr: true,
		},
		{
			name:    "no files",
			files:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean cache directory
			files, _ := filepath.Glob(filepath.Join(cacheDir, "*"))
			for _, f := range files {
				os.Remove(f)
			}

			for filename, content := range tt.files {
				path := filepath.Join(cacheDir, filename)
				err := os.WriteFile(path, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			err := FindValidSession()
			if (err != nil) != tt.wantErr {
				t.Errorf("FindValidSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
