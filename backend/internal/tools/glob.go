package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type GlobTool struct {
	workingDir string
}

func NewGlobTool(workingDir string) *GlobTool {
	return &GlobTool{
		workingDir: workingDir,
	}
}

func (t *GlobTool) Name() string {
	return "glob"
}

func (t *GlobTool) Description() string {
	return "Find files matching a glob pattern. Supports ** for recursive matching (e.g., '**/*.go', 'src/**/*.ts')"
}

func (t *GlobTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to match files (e.g., '**/*.go', 'src/**/*.ts')",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in (defaults to working directory)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("pattern is required")
	}

	searchPath := t.workingDir
	if path, ok := args["path"].(string); ok {
		if !filepath.IsAbs(path) {
			searchPath = filepath.Join(t.workingDir, path)
		} else {
			searchPath = path
		}
	}

	searchPath = filepath.Clean(searchPath)
	if !strings.HasPrefix(searchPath, t.workingDir+string(filepath.Separator)) && searchPath != t.workingDir {
		return ToolResult{}, fmt.Errorf("access denied: path is outside working directory")
	}

	info, err := os.Stat(searchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolResult{}, fmt.Errorf("path not found: %s", searchPath)
		}
		return ToolResult{}, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return ToolResult{}, fmt.Errorf("path is not a directory: %s", searchPath)
	}

	var matches []string
	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(searchPath, path)
		if err != nil {
			return nil
		}

		relPath = filepath.ToSlash(relPath)

		if matchGlob(pattern, relPath) {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to walk directory: %w", err)
	}

	sort.Strings(matches)

	maxResults := 1000
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	var builder strings.Builder
	builder.WriteString("<glob_results>\n")

	for _, match := range matches {
		relPath, err := filepath.Rel(t.workingDir, match)
		if err != nil {
			relPath = match
		}
		builder.WriteString(relPath + "\n")
	}

	builder.WriteString(fmt.Sprintf("\nFound %d matches", len(matches)))
	if len(matches) == maxResults {
		builder.WriteString(" (showing first 1000 results)")
	}
	builder.WriteString("\n</glob_results>")

	return ToolResult{
		Content: builder.String(),
		IsError: false,
	}, nil
}

func matchGlob(pattern, path string) bool {
	if !strings.Contains(pattern, "**") {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	return matchGlobRecursive(patternParts, pathParts)
}

func matchGlobRecursive(patternParts, pathParts []string) bool {
	if len(patternParts) == 0 && len(pathParts) == 0 {
		return true
	}
	if len(patternParts) == 0 {
		return false
	}
	if len(pathParts) == 0 {
		for _, part := range patternParts {
			if part != "**" {
				return false
			}
		}
		return true
	}

	patternPart := patternParts[0]
	pathPart := pathParts[0]

	if patternPart == "**" {
		if matchGlobRecursive(patternParts[1:], pathParts) {
			return true
		}
		if matchGlobRecursive(patternParts, pathParts[1:]) {
			return true
		}
		return false
	}

	matched, _ := filepath.Match(patternPart, pathPart)
	if matched {
		return matchGlobRecursive(patternParts[1:], pathParts[1:])
	}

	return false
}