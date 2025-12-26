package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jack/klaudkod/backend/internal/api"
	"github.com/jack/klaudkod/backend/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Create WebSocket hub
	hub := api.NewHub(cfg)
	go hub.Run()

	// Setup HTTP server
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		api.ServeWs(hub, w, r)
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Klaudkod server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
		os.Exit(1)
	}
}
