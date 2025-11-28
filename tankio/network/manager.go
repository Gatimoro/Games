package network

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	MaxLobbies       = 5
	LobbyTimeout     = 30 * time.Minute
	CleanupInterval  = 1 * time.Minute
	LobbyCodeLength  = 4
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// LobbyManager manages all active game lobbies
type LobbyManager struct {
	lobbies map[string]*Lobby
	mu      sync.RWMutex
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies: make(map[string]*Lobby),
	}
}

// CreateLobby creates a new lobby and returns its code
func (lm *LobbyManager) CreateLobby() (*Lobby, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if len(lm.lobbies) >= MaxLobbies {
		return nil, fmt.Errorf("maximum number of lobbies reached (%d)", MaxLobbies)
	}

	// Generate unique code
	var code string
	for {
		code = generateLobbyCode()
		if _, exists := lm.lobbies[code]; !exists {
			break
		}
	}

	lobby := NewLobby(code, lm)
	lm.lobbies[code] = lobby

	// Start lobby goroutine
	go lobby.Run()

	log.Printf("Created lobby %s (%d/%d active)", code, len(lm.lobbies), MaxLobbies)
	return lobby, nil
}

// GetLobby returns a lobby by its code
func (lm *LobbyManager) GetLobby(code string) *Lobby {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.lobbies[code]
}

// RemoveLobby removes a lobby from the manager
func (lm *LobbyManager) RemoveLobby(code string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lobby, exists := lm.lobbies[code]; exists {
		lobby.Close()
		delete(lm.lobbies, code)
		log.Printf("Removed lobby %s (%d/%d active)", code, len(lm.lobbies), MaxLobbies)
	}
}

// CleanupRoutine periodically removes inactive lobbies
func (lm *LobbyManager) CleanupRoutine() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		lm.cleanupInactiveLobbies()
	}
}

func (lm *LobbyManager) cleanupInactiveLobbies() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	for code, lobby := range lm.lobbies {
		if now.Sub(lobby.LastActivity()) > LobbyTimeout {
			lobby.Close()
			delete(lm.lobbies, code)
			log.Printf("Cleaned up inactive lobby %s", code)
		}
	}
}

// HandleCreateLobby handles the POST /api/create-lobby endpoint
func (lm *LobbyManager) HandleCreateLobby(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lobby, err := lm.CreateLobby()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code": lobby.Code,
	})
}

// HandleListLobbies handles the GET /api/lobbies endpoint
func (lm *LobbyManager) HandleListLobbies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lobbies := make([]map[string]interface{}, 0)
	for code, lobby := range lm.lobbies {
		lobbies = append(lobbies, map[string]interface{}{
			"code":        code,
			"playerCount": lobby.GetPlayerCount(),
			"maxPlayers":  MaxPlayersPerLobby,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lobbies)
}

// HandleWebSocket handles WebSocket connections
func (lm *LobbyManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	lobbyCode := r.URL.Query().Get("lobby")
	if lobbyCode == "" {
		http.Error(w, "Missing lobby code", http.StatusBadRequest)
		return
	}

	lobby := lm.GetLobby(lobbyCode)
	if lobby == nil {
		http.Error(w, "Lobby not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Generate player ID
	playerID := generatePlayerID()

	client := NewClient(playerID, conn, lobby)

	// Register client with lobby
	lobby.Register <- client

	// Start client pumps
	go client.WritePump()
	go client.ReadPump()
}

// generateLobbyCode creates a short, easy-to-type lobby code
func generateLobbyCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Avoid ambiguous chars like 0, O, 1, I
	b := make([]byte, LobbyCodeLength)
	rand.Read(b)
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// generatePlayerID creates a unique player identifier
func generatePlayerID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("P%X", b)
}
