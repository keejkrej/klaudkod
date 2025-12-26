package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteFileTool struct {
	workingDir string
}

func NewWriteFileTool(workingDir string) *WriteFileTool {
	return &WriteFileTool{
		workingDir: workingDir,
	}
}

func (t *WriteFileTool) Name() string {
	return "write"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file, creating it if it doesn't exist or overwriting if it does. Supports creating parent directories as needed."
}

func (t *WriteFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"filePath": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write to the file",
			},
		},
		"required": []string{"filePath", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("filePath is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("content is required")
	}

	// Resolve to absolute path
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(t.workingDir, filePath)
	}

	// Clean the path
	filePath = filepath.Clean(filePath)

	// Validate path is within workingDir
	if !strings.HasPrefix(filePath, t.workingDir+string(filepath.Separator)) && filePath != t.workingDir {
		return ToolResult{}, fmt.Errorf("access denied: path is outside working directory")
	}

	// Check if file exists to determine if we're creating or overwriting
	_, err := os.Stat(filePath)
	fileExists := err == nil

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ToolResult{}, fmt.Errorf("failed to create directories: %w", err)
	}

	// Write the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Prepare success message
	action := "created"
	if fileExists {
		action = "overwritten"
	}

	message := fmt.Sprintf("File %s successfully (%d bytes written)", action, len(content))

	return ToolResult{
		Content: message,
		IsError: false,
	}, nil
}