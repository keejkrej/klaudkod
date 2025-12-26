package tools

import "context"

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (ToolResult, error)
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

type ToolContext struct {
	SessionID  string
	WorkingDir string
	AbortChan  chan struct{}
}

type PermissionMode string

const (
	PermissionModeAsk  PermissionMode = "ask"
	PermissionModeAuto PermissionMode = "auto"
)