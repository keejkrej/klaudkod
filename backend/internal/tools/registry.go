package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"
)

type Registry struct {
	tools          map[string]Tool
	workingDir     string
	permissionMode PermissionMode
}

func NewRegistry(workingDir string, mode PermissionMode) *Registry {
	return &Registry{
		tools:          make(map[string]Tool),
		workingDir:     workingDir,
		permissionMode: mode,
	}
}

func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

func (r *Registry) GetOpenAITools() []openai.ChatCompletionToolParam {
	var tools []openai.ChatCompletionToolParam
	
	for _, tool := range r.tools {
		tools = append(tools, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name(),
				Description: openai.String(tool.Description()),
				Parameters:  shared.FunctionParameters(tool.Parameters()),
			},
		})
	}
	
	return tools
}

func (r *Registry) Execute(ctx context.Context, name string, argsJSON string) (ToolResult, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ToolResult{}, fmt.Errorf("failed to parse arguments: %w", err)
	}
	
	tool, exists := r.Get(name)
	if !exists {
		return ToolResult{}, fmt.Errorf("tool '%s' not found", name)
	}
	
	result, err := tool.Execute(ctx, args)
	if err != nil {
		return ToolResult{}, fmt.Errorf("tool execution failed: %w", err)
	}
	
	return result, nil
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}