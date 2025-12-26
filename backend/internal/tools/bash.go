package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type BashTool struct {
	workingDir      string
	defaultTimeout  time.Duration
	maxOutputLength int
}

func NewBashTool(workingDir string) *BashTool {
	return &BashTool{
		workingDir:      workingDir,
		defaultTimeout:  2 * time.Minute,
		maxOutputLength: 30000,
	}
}

func (b *BashTool) Name() string {
	return "bash"
}

func (b *BashTool) Description() string {
	return "Execute shell commands with optional timeout and working directory. Supports running any shell command with configurable timeout (default 2 minutes) and custom working directory."
}

func (b *BashTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Optional timeout in milliseconds",
			},
			"workdir": map[string]interface{}{
				"type":        "string",
				"description": fmt.Sprintf("The working directory to run the command in. Defaults to %s. Use this instead of 'cd' commands.", b.workingDir),
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Clear, concise description of what this command does in 5-10 words. Examples:\nInput: ls\nOutput: Lists files in current directory\n\nInput: git status\nOutput: Shows working tree status\n\nInput: npm install\nOutput: Installs package dependencies\n\nInput: mkdir foo\nOutput: Creates directory 'foo'",
			},
		},
		"required": []string{"command", "description"},
	}
}

func (b *BashTool) Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error) {
	command, ok := args["command"].(string)
	if !ok {
		return ToolResult{}, fmt.Errorf("command is required and must be a string")
	}

	// description is passed but we just validate it exists
	if _, ok := args["description"].(string); !ok {
		return ToolResult{}, fmt.Errorf("description is required and must be a string")
	}

	timeout := b.defaultTimeout
	if timeoutMs, exists := args["timeout"]; exists {
		if tm, ok := timeoutMs.(float64); ok {
			timeout = time.Duration(tm) * time.Millisecond
		}
	}

	workdir := b.workingDir
	if wd, exists := args["workdir"]; exists {
		if dir, ok := wd.(string); ok && dir != "" {
			workdir = dir
		}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	cmd.Dir = workdir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	if stderr.Len() > 0 {
		output += "[stderr]" + stderr.String()
	}

	var metadata []string
	metadata = append(metadata, "<bash_metadata>")

	if len(output) > b.maxOutputLength {
		output = output[:b.maxOutputLength]
		metadata = append(metadata, fmt.Sprintf("bash tool truncated output as it exceeded %d char limit", b.maxOutputLength))
	}

	if cmdCtx.Err() == context.DeadlineExceeded {
		metadata = append(metadata, fmt.Sprintf("bash tool terminated command after exceeding timeout %v", timeout))
	}

	if len(metadata) > 1 {
		metadata = append(metadata, "</bash_metadata>")
		output += "\n\n" + strings.Join(metadata, "\n")
	}

	result := ToolResult{
		Content: output,
		IsError: err != nil,
	}

	if err != nil {
		result.Content = fmt.Sprintf("Command failed: %v\n\n%s", err, result.Content)
	}

	return result, nil
}