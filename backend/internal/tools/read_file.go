package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultReadLimit = 2000
	maxLineLength    = 2000
)

type ReadFileTool struct {
	workingDir string
}

func NewReadFileTool(workingDir string) *ReadFileTool {
	return &ReadFileTool{
		workingDir: workingDir,
	}
}

func (t *ReadFileTool) Name() string {
	return "read"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file. Supports pagination with offset and limit parameters. Returns file content with line numbers."
}

func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"filePath": map[string]interface{}{
				"type":        "string",
				"description": "The path to the file to read",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "The line number to start reading from (0-based)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "The number of lines to read (defaults to 2000)",
			},
		},
		"required": []string{"filePath"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("filePath is required")
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

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{}, fmt.Errorf("file not found: %s", filePath)
		}
		return ToolResult{}, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return ToolResult{}, fmt.Errorf("cannot read directory: %s", filePath)
	}

	// Block .env files (except whitelisted ones)
	basename := filepath.Base(filePath)
	basenameLower := strings.ToLower(basename)

	// Whitelist for example/template files
	whitelist := []string{".env.sample", ".env.example", ".env.template", ".example"}
	isWhitelisted := false
	for _, w := range whitelist {
		if strings.HasSuffix(basenameLower, w) {
			isWhitelisted = true
			break
		}
	}

	// Block .env files: those starting with ".env" or ending with ".env"
	isEnvFile := false
	if strings.HasPrefix(basenameLower, ".env") && (len(basename) == 4 || basename[4] == '.') {
		// Matches: .env, .env.local, .env.production, etc.
		isEnvFile = true
	} else if strings.HasSuffix(basenameLower, ".env") {
		// Matches: production.env, local.env, credentials.env, etc.
		isEnvFile = true
	}

	if !isWhitelisted && isEnvFile {
		return ToolResult{}, fmt.Errorf("access denied: cannot read .env files")
	}

	// Check for binary files by extension
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExts := map[string]bool{
		".zip": true, ".tar": true, ".gz": true, ".exe": true, ".dll": true,
		".so": true, ".class": true, ".jar": true, ".war": true, ".7z": true,
		".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true,
		".pptx": true, ".odt": true, ".ods": true, ".odp": true, ".bin": true,
		".dat": true, ".obj": true, ".o": true, ".a": true, ".lib": true,
		".wasm": true, ".pyc": true, ".pyo": true,
	}
	if binaryExts[ext] {
		return ToolResult{}, fmt.Errorf("cannot read binary file: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Check for null bytes (binary file indicator)
	if strings.Contains(string(content), "\x00") {
		return ToolResult{}, fmt.Errorf("cannot read binary file: %s", filePath)
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")

	// Get offset and limit
	offset := 0
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}
	limit := defaultReadLimit
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}
	if offset > len(lines) {
		offset = len(lines)
	}

	// Apply offset and limit
	end := offset + limit
	if end > len(lines) {
		end = len(lines)
	}
	selectedLines := lines[offset:end]

	// Format output with line numbers
	var builder strings.Builder
	builder.WriteString("<file>\n")

	for i, line := range selectedLines {
		lineNum := offset + i + 1
		if len(line) > maxLineLength {
			line = line[:maxLineLength] + "..."
		}
		builder.WriteString(fmt.Sprintf("%05d| %s\n", lineNum, line))
	}

	// Add file end information
	totalLines := len(lines)
	lastReadLine := offset + len(selectedLines)
	if lastReadLine < totalLines {
		builder.WriteString(fmt.Sprintf("\n(File has more lines. Use 'offset' parameter to read beyond line %d)\n", lastReadLine))
	} else {
		builder.WriteString(fmt.Sprintf("\n(End of file - total %d lines)\n", totalLines))
	}
	builder.WriteString("</file>")

	return ToolResult{
		Content: builder.String(),
		IsError: false,
	}, nil
}