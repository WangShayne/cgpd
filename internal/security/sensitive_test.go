package security

import (
	"testing"

	"cgpd/internal/config"
)

func TestCheckSensitiveFiles(t *testing.T) {
	tests := []struct {
		name          string
		cfg           config.SecurityConfig
		files         []string
		expectedCount int
		expectedFiles []string
	}{
		{
			name: "detect .env file",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{"main.go", ".env", "README.md"},
			expectedCount: 1,
			expectedFiles: []string{".env"},
		},
		{
			name: "detect multiple sensitive files",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{".env", "id_rsa", "main.go", "config.pem"},
			expectedCount: 3,
			expectedFiles: []string{".env", "id_rsa", "config.pem"},
		},
		{
			name: "detect .env.* patterns",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{".env.local", ".env.production", "main.go"},
			expectedCount: 2,
			expectedFiles: []string{".env.local", ".env.production"},
		},
		{
			name: "case insensitive matching",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{".ENV", "SECRET.txt", "Password.key"},
			expectedCount: 3,
			expectedFiles: []string{".ENV", "SECRET.txt", "Password.key"},
		},
		{
			name: "detect files in subdirectories",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{"config/.env", "keys/id_rsa", "src/main.go"},
			expectedCount: 2,
			expectedFiles: []string{"config/.env", "keys/id_rsa"},
		},
		{
			name: "exclude patterns work",
			cfg: config.SecurityConfig{
				Enabled:         true,
				ExcludePatterns: []string{"example.env", "*.template"},
			},
			files:         []string{".env", "example.env", "secret.template", "id_rsa"},
			expectedCount: 2,
			expectedFiles: []string{".env", "id_rsa"},
		},
		{
			name: "additional patterns work",
			cfg: config.SecurityConfig{
				Enabled:            true,
				AdditionalPatterns: []string{"*.toml", "database.yml"},
			},
			files:         []string{"config.toml", "database.yml", "main.go"},
			expectedCount: 2,
			expectedFiles: []string{"config.toml", "database.yml"},
		},
		{
			name: "disabled security check returns empty",
			cfg: config.SecurityConfig{
				Enabled: false,
			},
			files:         []string{".env", "id_rsa", "secret.key"},
			expectedCount: 0,
			expectedFiles: []string{},
		},
		{
			name: "no sensitive files",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{"main.go", "README.md", "package.json"},
			expectedCount: 0,
			expectedFiles: []string{},
		},
		{
			name: "empty file list",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{},
			expectedCount: 0,
			expectedFiles: []string{},
		},
		{
			name: "detect .config.yaml and .cgpd.yaml",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{".config.yaml", ".cgpd.yaml", "main.go"},
			expectedCount: 2,
			expectedFiles: []string{".config.yaml", ".cgpd.yaml"},
		},
		{
			name: "wildcard patterns with *",
			cfg: config.SecurityConfig{
				Enabled: true,
			},
			files:         []string{"api_key.txt", "my_secret.conf", "user_password.dat"},
			expectedCount: 3,
			expectedFiles: []string{"api_key.txt", "my_secret.conf", "user_password.dat"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CheckSensitiveFiles(tt.cfg, tt.files)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tt.expectedCount {
				t.Errorf("expected %d sensitive files, got %d: %v", tt.expectedCount, len(result), result)
			}

			// Verify each expected file is in the result
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, resultFile := range result {
					if resultFile == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected file %s not found in results: %v", expectedFile, result)
				}
			}
		})
	}
}

func TestIsSensitiveFile(t *testing.T) {
	patterns := []string{".env", "*.key", "*secret*"}
	excludePatterns := []string{"example.env"}

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"exact match", ".env", true},
		{"wildcard match", "api.key", true},
		{"contains match", "my_secret.txt", true},
		{"excluded file", "example.env", false},
		{"normal file", "main.go", false},
		{"path with sensitive basename", "config/.env", true},
		{"case insensitive", ".ENV", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSensitiveFile(tt.filename, patterns, excludePatterns)
			if result != tt.expected {
				t.Errorf("isSensitiveFile(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		pattern  string
		expected bool
	}{
		{"exact match", ".env", ".env", true},
		{"wildcard prefix", "test.key", "*.key", true},
		{"wildcard suffix", "secret_file", "*secret*", true},
		{"no match", "main.go", "*.key", false},
		{"invalid pattern", "test", "[invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.filename, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.filename, tt.pattern, result, tt.expected)
			}
		})
	}
}
