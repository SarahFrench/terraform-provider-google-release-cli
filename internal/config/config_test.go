package config

import (
	"fmt"
	"os"
	"testing"
)

func TestConfig_validate(t *testing.T) {

	// A directory to use as a value
	tmpDir := os.TempDir()
	nonExistentDir := "/User/doesntexist/path/to/nowhere"

	testCases := map[string]struct {
		config      *Config
		expectError bool
	}{
		"all fields set": {
			config: &Config{
				MagicModulesPath: tmpDir,
				GooglePath:       tmpDir,
				GoogleBetaPath:   tmpDir,
				Remote:           tmpDir,
			},
		},
		"MagicModulesPath unset": {
			expectError: true,
			config: &Config{
				GooglePath:     tmpDir,
				GoogleBetaPath: tmpDir,
				Remote:         tmpDir,
			},
		},
		"MagicModulesPath bad path": {
			expectError: true,
			config: &Config{
				MagicModulesPath: nonExistentDir,
				GooglePath:       tmpDir,
				GoogleBetaPath:   tmpDir,
				Remote:           tmpDir,
			},
		},
		"GooglePath unset": {
			expectError: true,
			config: &Config{
				MagicModulesPath: tmpDir,
				GoogleBetaPath:   tmpDir,
				Remote:           tmpDir,
			},
		},
		"GooglePath bad path": {
			expectError: true,
			config: &Config{
				MagicModulesPath: tmpDir,
				GooglePath:       nonExistentDir,
				GoogleBetaPath:   tmpDir,
				Remote:           tmpDir,
			},
		},
		"GoogleBetaPath unset": {
			expectError: true,
			config: &Config{
				MagicModulesPath: tmpDir,
				GooglePath:       tmpDir,
				Remote:           tmpDir,
			},
		},
		"GoogleBetaPath bad path": {
			expectError: true,
			config: &Config{
				MagicModulesPath: tmpDir,
				GooglePath:       tmpDir,
				GoogleBetaPath:   nonExistentDir,
				Remote:           tmpDir,
			},
		},
		"Remote unset": {
			expectError: true,
			config: &Config{
				MagicModulesPath: tmpDir,
				GooglePath:       tmpDir,
				GoogleBetaPath:   tmpDir,
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			tc.config.RemoteOwner = "hashicorp" // this is the default, added when creating configs via constructor
			err := tc.config.validate()
			if err != nil {
				if tc.expectError == false {
					t.Fatalf("unexpected error(s) encountered: %v", err)
				}
				// error is expected but we don't assert what error
			}
		})
	}
}

// TestConfig_LoadConfigFromFile creates a config file in a temporary directory that has valid values inside.
// The test checks that LoadConfigFromFile can find the file and read the values into the struct as expected,
// without error
func TestConfig_LoadConfigFromFile(t *testing.T) {

	// Arrange
	tmpDir := os.TempDir()
	remote := "origin"
	owner := "sarahcorp"

	json := fmt.Sprintf(`{
	"magicModulesPath": "%s",
	"googlePath": "%s",
	"googleBetaPath": "%s",
	"remote": "%s",
	"remoteOwner": "%s"
}`, tmpDir, tmpDir, tmpDir, remote, owner) // paths are all valid

	// Make a test fixture that contains paths to existing directories
	f, err := os.Create(tmpDir + CONFIG_FILE_NAME)
	if err != nil {
		t.Fatalf("error creating temporary %s file: %s", CONFIG_FILE_NAME, err)
	}
	_, err = f.Write([]byte(json))
	if err != nil {
		t.Fatalf("error writing to temporary %s file: %s", CONFIG_FILE_NAME, err)
	}
	err = f.Close()
	if err != nil {
		t.Fatalf("error closing temporary %s file: %s", CONFIG_FILE_NAME, err)
	}

	// The function under test looks for a file called .tpg-cli-config.json located at $HOME
	// so we need to temporarily make HOME the path to our temp folder.
	t.Setenv("HOME", tmpDir)

	// Act
	c, err := LoadConfigFromFile()

	// Assert
	if err != nil {
		t.Fatalf("unexpected error(s) encountered: %s", err)
	}

	if c.MagicModulesPath != tmpDir {
		t.Fatalf("unexpected value of MagicModulesPath, want %s, got %s", tmpDir, c.MagicModulesPath)
	}
	if c.GooglePath != tmpDir {
		t.Fatalf("unexpected value of GooglePath, want %s, got %s", tmpDir, c.GooglePath)
	}
	if c.GoogleBetaPath != tmpDir {
		t.Fatalf("unexpected value of GooglePath, want %s, got %s", tmpDir, c.GoogleBetaPath)
	}
	if c.Remote != remote {
		t.Fatalf("unexpected value of Remote, want %s, got %s", remote, c.Remote)
	}
	if c.RemoteOwner != owner {
		t.Fatalf("unexpected value of RemoteOwner, want %s, got %s", owner, c.RemoteOwner)
	}
}
