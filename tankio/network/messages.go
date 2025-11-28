package network

import (
	"tankio/game"
)

// ClientMessage represents a message from client to server
type ClientMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Message types from client
const (
	MsgTypeInput        = "input"         // Player input state
	MsgTypeFire         = "fire"          // Fire weapon
	MsgTypeSwitchWeapon = "switch_weapon" // Change weapon
)

// InputPayload is the payload for input messages
type InputPayload struct {
	Up     bool    `json:"up"`
	Down   bool    `json:"down"`
	Left   bool    `json:"left"`
	Right  bool    `json:"right"`
	MouseX float64 `json:"mouseX"`
	MouseY float64 `json:"mouseY"`
	Firing bool    `json:"firing"`
}

// SwitchWeaponPayload is the payload for weapon switch messages
type SwitchWeaponPayload struct {
	Weapon game.WeaponType `json:"weapon"`
}

// ServerMessage represents a message from server to client
type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Message types from server
const (
	MsgTypeGameState    = "game_state"    // Full game state update
	MsgTypePlayerJoined = "player_joined" // A player joined
	MsgTypePlayerLeft   = "player_left"   // A player left
	MsgTypeError        = "error"         // Error message
	MsgTypeConnected    = "connected"     // Connection confirmed
	MsgTypeLobbyInfo    = "lobby_info"    // Lobby information
)

// ConnectedPayload is sent when a player successfully connects
type ConnectedPayload struct {
	PlayerID  string `json:"playerId"`
	LobbyCode string `json:"lobbyCode"`
}

// LobbyInfoPayload contains lobby state
type LobbyInfoPayload struct {
	Code        string   `json:"code"`
	PlayerCount int      `json:"playerCount"`
	MaxPlayers  int      `json:"maxPlayers"`
	State       string   `json:"state"`
	Players     []string `json:"players"`
}

// ErrorPayload contains error information
type ErrorPayload struct {
	Message string `json:"message"`
}
