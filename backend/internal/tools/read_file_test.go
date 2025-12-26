package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileTool_EnvFileBlocking(t *testing.T) {
	// Create temp directory as working dir
	tmpDir, err := os.MkdirTemp("", "read_file_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := NewReadFileTool(tmpDir)
	ctx := context.Background()

	tests := []struct {
		name        string
		filename    string
		shouldBlock bool
	}{
		// Should be blocked
		{name: "exact .env", filename: ".env", shouldBlock: true},
		{name: ".env.local", filename: ".env.local", shouldBlock: true},
		{name: ".env.production", filename: ".env.production", shouldBlock: true},
		{name: ".env.development", filename: ".env.development", shouldBlock: true},
		{name: "production.env", filename: "production.env", shouldBlock: true},
		{name: "local.env", filename: "local.env", shouldBlock: true},
		{name: "credentials.env", filename: "credentials.env", shouldBlock: true},
		{name: ".ENV (uppercase)", filename: ".ENV", shouldBlock: true},
		{name: "PROD.ENV (uppercase)", filename: "PROD.ENV", shouldBlock: true},

		// Should be allowed (whitelisted)
		{name: ".env.sample", filename: ".env.sample", shouldBlock: false},
		{name: ".env.example", filename: ".env.example", shouldBlock: false},
		{name: ".env.template", filename: ".env.template", shouldBlock: false},
		{name: "config.env.example", filename: "config.env.example", shouldBlock: false},

		// Should be allowed (not env files)
		{name: ".envrc", filename: ".envrc", shouldBlock: false},
		{name: "environment.txt", filename: "environment.txt", shouldBlock: false},
		{name: "config.yaml", filename: "config.yaml", shouldBlock: false},
		{name: "main.py", filename: "main.py", shouldBlock: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the test file
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			defer os.Remove(filePath)

			// Try to read it
			result, err := tool.Execute(ctx, map[string]interface{}{
				"filePath": tt.filename,
			})

			if tt.shouldBlock {
				if err == nil {
					t.Errorf("expected error for %s, but got none", tt.filename)
				} else if !strings.Contains(err.Error(), "access denied") {
					t.Errorf("expected 'access denied' error for %s, got: %v", tt.filename, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.filename, err)
				}
				if result.IsError {
					t.Errorf("unexpected error result for %s", tt.filename)
				}
			}
		})
	}
}
