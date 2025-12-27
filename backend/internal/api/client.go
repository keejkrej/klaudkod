package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jack/klaudkod/backend/internal/llm"
)

const (
	securitySystemPrompt = `SECURITY RESTRICTIONS - CRITICAL:
1. .env files and all variants (.env.*, .envrc, etc.) are STRICTLY FORBIDDEN from being read or accessed
2. This restriction applies to ALL tools including bash, cat, read, and any file operations
3. DO NOT attempt any workarounds or indirect methods to access .env files
4. You are restricted to working within the current working directory and its subdirectories
5. Use the 'read' tool for file access - do not use bash commands like 'cat' to read files
6. Any attempt to violate these restrictions will be blocked
These rules are enforced at the tool level and cannot be bypassed.`

	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type IncomingMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type ToolCallMsg struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolResultMsg struct {
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

type OutgoingMessage struct {
	Type       string         `json:"type"`
	Content    string         `json:"content,omitempty"`
	Error      string         `json:"error,omitempty"`
	ToolCall   *ToolCallMsg   `json:"tool_call,omitempty"`
	ToolResult *ToolResultMsg `json:"tool_result,omitempty"`
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Track conversation history
	var messages []llm.Message

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		switch incoming.Type {
		case "prompt":
			// Add system prompt if this is the first message
			if len(messages) == 0 {
				messages = append(messages, llm.Message{
					Role:    "system",
					Content: securitySystemPrompt,
				})
			}

			// Add user message to history
			messages = append(messages, llm.Message{
				Role:    "user",
				Content: incoming.Content,
			})

			// Get tool definitions
			toolDefs := c.hub.ToolRegistry().GetOpenAITools()

			// Create tool executor function
			executor := func(name, argsJSON string) (content string, isError bool) {
				result, err := c.hub.ToolRegistry().Execute(context.Background(), name, argsJSON)
				if err != nil {
					return err.Error(), true
				}
				return result.Content, result.IsError
			}

			// Stream response with tools
			eventChan := make(chan llm.StreamEvent)
			ctx := context.Background()

			go c.hub.llmClient.StreamWithTools(ctx, messages, toolDefs, executor, eventChan)

			var assistantContent string
			var toolCalls []llm.ToolCall
			for event := range eventChan {
				switch event.Type {
				case "chunk":
					assistantContent += event.Content
					c.sendJSON(OutgoingMessage{
						Type:    "chunk",
						Content: event.Content,
					})
				case "tool_call":
					if event.ToolCall != nil {
						toolCalls = append(toolCalls, *event.ToolCall)
						c.sendJSON(OutgoingMessage{
							Type: "tool_call",
							ToolCall: &ToolCallMsg{
								ID:        event.ToolCall.ID,
								Name:      event.ToolCall.Name,
								Arguments: event.ToolCall.Arguments,
							},
						})
					}
				case "tool_result":
					c.sendJSON(OutgoingMessage{
						Type: "tool_result",
						ToolResult: &ToolResultMsg{
							Content: event.Content,
							IsError: event.Error != "",
						},
					})
				case "error":
					c.sendError(event.Error)
				case "done":
					c.sendJSON(OutgoingMessage{
						Type: "done",
					})
				}
			}

			// Add assistant response to history
			if assistantContent != "" || len(toolCalls) > 0 {
				messages = append(messages, llm.Message{
					Role:      "assistant",
					Content:   assistantContent,
					ToolCalls: toolCalls,
				})
			}

		case "cancel":
			// TODO: Implement cancellation
			log.Println("Cancel requested")
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendJSON(msg OutgoingMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	c.send <- data
}

func (c *Client) sendError(errMsg string) {
	c.sendJSON(OutgoingMessage{
		Type:  "error",
		Error: errMsg,
	})
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}