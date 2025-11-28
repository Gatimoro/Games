package main

import (
	"log"
	"net/http"
	"tankio/network"
)

func main() {
	// Create the lobby manager (handles all game lobbies)
	lobbyManager := network.NewLobbyManager()

	// Start the lobby cleanup routine (kills inactive lobbies after 30 min)
	go lobbyManager.CleanupRoutine()

	// Serve static files (HTML, JS, CSS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// API endpoints for lobby management
	http.HandleFunc("/api/create-lobby", lobbyManager.HandleCreateLobby)
	http.HandleFunc("/api/lobbies", lobbyManager.HandleListLobbies)

	// WebSocket endpoint for game connections
	http.HandleFunc("/ws", lobbyManager.HandleWebSocket)

	log.Println("🎮 Tank.io server starting on http://localhost:8080")
	log.Println("   Open your browser and navigate to the URL above")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
