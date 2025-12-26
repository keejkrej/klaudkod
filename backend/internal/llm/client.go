package llm

import (
	"context"
	"strings"

	"github.com/jack/klaudkod/backend/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	client openai.Client
	model  string
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string    `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type StreamEvent struct {
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	Error     string    `json:"error,omitempty"`
	ToolCall  *ToolCall `json:"tool_call,omitempty"`
}

type ToolExecutor func(name, argsJSON string) (content string, isError bool)

func NewClient(cfg *config.Config) *Client {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.LLMAPIKey),
	}

	if cfg.LLMBaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.LLMBaseURL))
	}

	client := openai.NewClient(opts...)

	return &Client{
		client: client,
		model:  cfg.LLMModel,
	}
}

func (c *Client) Stream(ctx context.Context, messages []Message, eventChan chan<- StreamEvent) {
	defer close(eventChan)

	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			openaiMessages[i] = openai.AssistantMessage(msg.Content)
		case "system":
			openaiMessages[i] = openai.SystemMessage(msg.Content)
		default:
			openaiMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	// Create streaming request
	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(c.model),
		Messages: openaiMessages,
	})

	for stream.Next() {
		chunk := stream.Current()
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				eventChan <- StreamEvent{
					Type:    "chunk",
					Content: choice.Delta.Content,
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		eventChan <- StreamEvent{
			Type:  "error",
			Error: err.Error(),
		}
		return
	}

	eventChan <- StreamEvent{
		Type: "done",
	}
}

func (c *Client) StreamWithTools(ctx context.Context, messages []Message, tools []openai.ChatCompletionToolParam, executor ToolExecutor, eventChan chan<- StreamEvent) {
	defer close(eventChan)

	currentMessages := make([]Message, len(messages))
	copy(currentMessages, messages)

	for {
		// Convert messages to OpenAI format
		openaiMessages := c.convertMessagesToOpenAI(currentMessages)

		// Create streaming request with tools
		stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModel(c.model),
			Messages: openaiMessages,
			Tools:    tools,
		})

		var contentBuilder strings.Builder
		var toolCalls []ToolCall
		var currentToolCall *ToolCall
		var currentArgsBuilder strings.Builder

		for stream.Next() {
			chunk := stream.Current()
			for _, choice := range chunk.Choices {
				// Handle content chunks
				if choice.Delta.Content != "" {
					contentBuilder.WriteString(choice.Delta.Content)
					eventChan <- StreamEvent{
						Type:    "chunk",
						Content: choice.Delta.Content,
					}
				}

				// Handle tool calls
				for _, deltaToolCall := range choice.Delta.ToolCalls {
					// Initialize new tool call if needed
					if currentToolCall == nil || currentToolCall.ID != deltaToolCall.ID {
						if currentToolCall != nil && currentToolCall.Arguments != "" {
							toolCalls = append(toolCalls, *currentToolCall)
						}
						currentToolCall = &ToolCall{
							ID:   deltaToolCall.ID,
							Name: deltaToolCall.Function.Name,
						}
						currentArgsBuilder.Reset()
					}

					// Update tool call name if provided
					if deltaToolCall.Function.Name != "" {
						currentToolCall.Name = deltaToolCall.Function.Name
					}

					// Accumulate arguments
					if deltaToolCall.Function.Arguments != "" {
						currentArgsBuilder.WriteString(deltaToolCall.Function.Arguments)
						currentToolCall.Arguments = currentArgsBuilder.String()
					}
				}
			}
		}

		// Add the last tool call if exists
		if currentToolCall != nil && currentToolCall.Arguments != "" {
			toolCalls = append(toolCalls, *currentToolCall)
		}

		if err := stream.Err(); err != nil {
			eventChan <- StreamEvent{
				Type:  "error",
				Error: err.Error(),
			}
			return
		}

		// Add assistant message to history
		assistantMsg := Message{
			Role:      "assistant",
			Content:   contentBuilder.String(),
			ToolCalls: toolCalls,
		}
		currentMessages = append(currentMessages, assistantMsg)

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			eventChan <- StreamEvent{
				Type: "done",
			}
			return
		}

		// Execute tools and add responses
		for _, toolCall := range toolCalls {
			// Emit tool call event
			eventChan <- StreamEvent{
				Type:     "tool_call",
				ToolCall: &toolCall,
			}

			// Execute tool
			content, isError := executor(toolCall.Name, toolCall.Arguments)

			// Emit tool result event
			eventChan <- StreamEvent{
				Type:    "tool_result",
				Content: content,
				Error:   func() string { if isError { return "error" } else { return "" } }(),
			}

			// Add tool response message
			toolMsg := Message{
				Role:       "tool",
				Content:    content,
				ToolCallID: toolCall.ID,
			}
			currentMessages = append(currentMessages, toolMsg)
		}
	}
}

func (c *Client) convertMessagesToOpenAI(messages []Message) []openai.ChatCompletionMessageParamUnion {
	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						},
					}
				}
				openaiMessages[i] = openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						ToolCalls: toolCalls,
					},
				}
			} else {
				openaiMessages[i] = openai.AssistantMessage(msg.Content)
			}
		case "system":
			openaiMessages[i] = openai.SystemMessage(msg.Content)
		case "tool":
			openaiMessages[i] = openai.ToolMessage(msg.Content, msg.ToolCallID)
		default:
			openaiMessages[i] = openai.UserMessage(msg.Content)
		}
	}
	return openaiMessages
}