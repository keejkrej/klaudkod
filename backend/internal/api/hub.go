package api

import (
	"os"
	"github.com/jack/klaudkod/backend/internal/config"
	"github.com/jack/klaudkod/backend/internal/llm"
	"github.com/jack/klaudkod/backend/internal/tools"
)

type Hub struct {
	config        *config.Config
	llmClient     *llm.Client
	clients       map[*Client]bool
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
	toolRegistry  *tools.Registry
}

func NewHub(cfg *config.Config) *Hub {
	workingDir, _ := os.Getwd()
	registry := tools.NewRegistry(workingDir, tools.PermissionModeAuto)
	registry.Register(tools.NewReadFileTool(workingDir))
	registry.Register(tools.NewWriteFileTool(workingDir))
	registry.Register(tools.NewGlobTool(workingDir))
	registry.Register(tools.NewGrepTool(workingDir))
	registry.Register(tools.NewBashTool(workingDir))
	
	return &Hub{
		config:       cfg,
		llmClient:    llm.NewClient(cfg),
		broadcast:    make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		clients:      make(map[*Client]bool),
		toolRegistry: registry,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) ToolRegistry() *tools.Registry {
	return h.toolRegistry
}