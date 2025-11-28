package network

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"tankio/game"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a connected player
type Client struct {
	ID       string
	Lobby    *Lobby
	Conn     *websocket.Conn
	Send     chan []byte
	done     chan struct{}
}

// NewClient creates a new client instance
func NewClient(id string, conn *websocket.Conn, lobby *Lobby) *Client {
	return &Client{
		ID:    id,
		Lobby: lobby,
		Conn:  conn,
		Send:  make(chan []byte, 256),
		done:  make(chan struct{}),
	}
}

// ReadPump pumps messages from the websocket connection to the lobby
func (c *Client) ReadPump() {
	defer func() {
		c.Lobby.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump pumps messages from the lobby to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Lobby closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Coalesce queued messages into single write
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

// handleMessage processes an incoming message from the client
func (c *Client) handleMessage(data []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch msg.Type {
	case MsgTypeInput:
		c.handleInput(msg.Payload)
	case MsgTypeFire:
		c.handleFire()
	case MsgTypeSwitchWeapon:
		c.handleSwitchWeapon(msg.Payload)
	}
}

func (c *Client) handleInput(payload interface{}) {
	data, _ := json.Marshal(payload)
	var input InputPayload
	if err := json.Unmarshal(data, &input); err != nil {
		return
	}

	c.Lobby.Game.HandleInput(game.PlayerInput{
		PlayerID: c.ID,
		Action:   "input",
		Input: game.InputState{
			Up:     input.Up,
			Down:   input.Down,
			Left:   input.Left,
			Right:  input.Right,
			MouseX: input.MouseX,
			MouseY: input.MouseY,
			Firing: input.Firing,
		},
	})
}

func (c *Client) handleFire() {
	c.Lobby.Game.HandleInput(game.PlayerInput{
		PlayerID: c.ID,
		Action:   "fire",
	})
}

func (c *Client) handleSwitchWeapon(payload interface{}) {
	data, _ := json.Marshal(payload)
	var wp SwitchWeaponPayload
	if err := json.Unmarshal(data, &wp); err != nil {
		return
	}

	c.Lobby.Game.HandleInput(game.PlayerInput{
		PlayerID: c.ID,
		Action:   "switch_weapon",
		Weapon:   wp.Weapon,
	})
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.Send <- data:
	default:
		// Buffer full, close connection
		close(c.done)
	}
}

// Close terminates the client connection
func (c *Client) Close() {
	close(c.done)
}
