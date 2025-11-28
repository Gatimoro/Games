package network

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"tankio/game"
)

const (
	MaxPlayersPerLobby = 2
)

// Lobby represents a game room that players can join
type Lobby struct {
	Code       string
	Game       *game.Game
	Clients    map[string]*Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte

	mu           sync.RWMutex
	lastActivity time.Time
	closed       bool
	manager      *LobbyManager
}

// NewLobby creates a new lobby with the given code
func NewLobby(code string, manager *LobbyManager) *Lobby {
	lobby := &Lobby{
		Code:         code,
		Game:         game.NewGame(),
		Clients:      make(map[string]*Client),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Broadcast:    make(chan []byte, 256),
		lastActivity: time.Now(),
		manager:      manager,
	}

	// Set up game broadcast function
	lobby.Game.BroadcastFn = func(msg interface{}) {
		lobby.BroadcastMessage(msg)
	}

	return lobby
}

// Run starts the lobby's main loop
func (l *Lobby) Run() {
	for {
		select {
		case client := <-l.Register:
			l.handleRegister(client)

		case client := <-l.Unregister:
			l.handleUnregister(client)

		case message := <-l.Broadcast:
			l.handleBroadcast(message)
		}

		// Check if lobby should be closed
		l.mu.RLock()
		closed := l.closed
		l.mu.RUnlock()
		if closed {
			return
		}
	}
}

func (l *Lobby) handleRegister(client *Client) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Clients) >= MaxPlayersPerLobby {
		client.SendMessage(ServerMessage{
			Type:    MsgTypeError,
			Payload: ErrorPayload{Message: "Lobby is full"},
		})
		client.Close()
		return
	}

	l.Clients[client.ID] = client
	l.lastActivity = time.Now()

	// Add player to game
	l.Game.AddPlayer(client.ID)

	// Send connection confirmation
	client.SendMessage(ServerMessage{
		Type: MsgTypeConnected,
		Payload: ConnectedPayload{
			PlayerID:  client.ID,
			LobbyCode: l.Code,
		},
	})

	// Send lobby info to all players
	l.broadcastLobbyInfo()

	log.Printf("Player %s joined lobby %s (%d/%d players)",
		client.ID, l.Code, len(l.Clients), MaxPlayersPerLobby)
}

func (l *Lobby) handleUnregister(client *Client) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.Clients[client.ID]; ok {
		delete(l.Clients, client.ID)
		close(client.Send)
		l.Game.RemovePlayer(client.ID)
		l.lastActivity = time.Now()

		log.Printf("Player %s left lobby %s", client.ID, l.Code)

		// If no players left, mark lobby for cleanup
		if len(l.Clients) == 0 {
			l.closed = true
			l.manager.RemoveLobby(l.Code)
		} else {
			l.broadcastLobbyInfo()
		}
	}
}

func (l *Lobby) handleBroadcast(message []byte) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, client := range l.Clients {
		select {
		case client.Send <- message:
		default:
			// Client buffer full, will be cleaned up
		}
	}
}

// BroadcastMessage sends a message to all clients
func (l *Lobby) BroadcastMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case l.Broadcast <- data:
	default:
		// Broadcast buffer full
	}
}

func (l *Lobby) broadcastLobbyInfo() {
	players := make([]string, 0, len(l.Clients))
	for id := range l.Clients {
		players = append(players, id)
	}

	l.BroadcastMessage(ServerMessage{
		Type: MsgTypeLobbyInfo,
		Payload: LobbyInfoPayload{
			Code:        l.Code,
			PlayerCount: len(l.Clients),
			MaxPlayers:  MaxPlayersPerLobby,
			State:       string(l.Game.State),
			Players:     players,
		},
	})
}

// GetPlayerCount returns the number of connected players
func (l *Lobby) GetPlayerCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Clients)
}

// LastActivity returns when the lobby was last active
func (l *Lobby) LastActivity() time.Time {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.lastActivity
}

// Close shuts down the lobby
func (l *Lobby) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}
	l.closed = true

	// Stop the game
	l.Game.Stop()

	// Close all client connections
	for _, client := range l.Clients {
		client.Close()
	}
}
