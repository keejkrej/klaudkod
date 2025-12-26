package llm

import (
	"context"

	"github.com/jack/klaudkod/backend/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Client struct {
	client openai.Client
	model  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StreamEvent struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

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
