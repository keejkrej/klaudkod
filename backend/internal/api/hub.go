package api

import (
	"github.com/jack/klaudkod/backend/internal/config"
	"github.com/jack/klaudkod/backend/internal/llm"
)

type Hub struct {
	config     *config.Config
	llmClient  *llm.Client
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func NewHub(cfg *config.Config) *Hub {
	return &Hub{
		config:     cfg,
		llmClient:  llm.NewClient(cfg),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
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
