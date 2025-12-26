package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GrepTool struct {
	workingDir string
	maxResults int
}

func NewGrepTool(workingDir string) *GrepTool {
	return &GrepTool{
		workingDir: workingDir,
		maxResults: 100,
	}
}

func (t *GrepTool) Name() string {
	return "grep"
}

func (t *GrepTool) Description() string {
	return "Search for regex patterns in file contents. Supports file inclusion patterns and line-by-line matching"
}

func (t *GrepTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory or file to search in (defaults to working directory)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern for files to include (e.g. '*.go')",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
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

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return ToolResult{}, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var includePattern string
	if include, ok := args["include"].(string); ok {
		includePattern = include
	}

	var matches []string
	var filesSearched int
	var totalMatches int

	skipDirs := map[string]bool{
		".git":        true,
		"node_modules": true,
		"vendor":      true,
		"__pycache__": true,
		".venv":       true,
	}

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if includePattern != "" {
			relPath, err := filepath.Rel(searchPath, path)
			if err != nil {
				return nil
			}
			relPath = filepath.ToSlash(relPath)
			if !matchGlob(includePattern, filepath.Base(relPath)) {
				return nil
			}
		}

		if t.isBinaryFile(path) {
			return nil
		}

		filesSearched++
		fileMatches, err := t.searchFile(path, regex)
		if err != nil {
			return nil
		}

		for _, match := range fileMatches {
			if totalMatches >= t.maxResults {
				return nil
			}
			matches = append(matches, match)
			totalMatches++
		}

		return nil
	})

	if err != nil {
		return ToolResult{}, fmt.Errorf("failed to walk directory: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("<grep_results>\n")

	for _, match := range matches {
		relPath, err := filepath.Rel(t.workingDir, strings.Split(match, ":")[0])
		if err != nil {
			relPath = strings.Split(match, ":")[0]
		}
		parts := strings.SplitN(match, ":", 3)
		if len(parts) == 3 {
			builder.WriteString(fmt.Sprintf("%s:%s:%s\n", relPath, parts[1], parts[2]))
		}
	}

	builder.WriteString(fmt.Sprintf("\nFound %d matches in %d files", totalMatches, filesSearched))
	if totalMatches == t.maxResults {
		builder.WriteString(fmt.Sprintf(" (showing first %d results)", t.maxResults))
	}
	builder.WriteString("\n</grep_results>")

	return ToolResult{
		Content: builder.String(),
		IsError: false,
	}, nil
}

func (t *GrepTool) isBinaryFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return true
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return true
	}

	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return true
		}
	}

	return false
}

func (t *GrepTool) searchFile(path string, regex *regexp.Regexp) ([]string, error) {
	var matches []string

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		if regex.MatchString(line) {
			matches = append(matches, fmt.Sprintf("%s:%d:%s", path, lineNum, line))
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}