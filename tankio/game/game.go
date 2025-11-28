package game

import (
	"sync"
	"time"
)

const (
	TickRate = 60 // Server updates per second
)

// GameState represents the current state of a game match
type GameState string

const (
	StateWaiting  GameState = "waiting"  // Waiting for players
	StatePlaying  GameState = "playing"  // Game in progress
	StateGameOver GameState = "gameover" // Game ended
)

// Game manages the game logic for a single match
type Game struct {
	mu sync.RWMutex

	State    GameState
	Map      MapConfig
	Tanks    map[string]*Tank
	Bullets  []*Bullet
	WinnerID string

	// Channels for communication
	InputChan    chan PlayerInput
	BroadcastFn  func(msg interface{})

	// Game loop control
	stopChan chan struct{}
	running  bool
}

// PlayerInput represents input received from a player
type PlayerInput struct {
	PlayerID string
	Input    InputState
	Action   string // "fire", "switch_weapon", etc.
	Weapon   WeaponType
}

// NewGame creates a new game instance
func NewGame() *Game {
	return &Game{
		State:     StateWaiting,
		Map:       DefaultMap(),
		Tanks:     make(map[string]*Tank),
		Bullets:   make([]*Bullet, 0),
		InputChan: make(chan PlayerInput, 100),
		stopChan:  make(chan struct{}),
	}
}

// AddPlayer adds a new player to the game
func (g *Game) AddPlayer(playerID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.Tanks) >= 2 {
		return false
	}

	spawnPoints := g.Map.GetSpawnPoints()
	spawnIndex := len(g.Tanks)
	tank := NewTank(playerID, spawnPoints[spawnIndex])
	g.Tanks[playerID] = tank

	// Start game when we have 2 players
	if len(g.Tanks) == 2 && g.State == StateWaiting {
		g.State = StatePlaying
		go g.Run()
	}

	return true
}

// RemovePlayer removes a player from the game
func (g *Game) RemovePlayer(playerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.Tanks, playerID)

	// If a player leaves during a game, the other player wins
	if g.State == StatePlaying && len(g.Tanks) == 1 {
		for id := range g.Tanks {
			g.WinnerID = id
		}
		g.State = StateGameOver
	}
}

// Run starts the game loop
func (g *Game) Run() {
	g.running = true
	ticker := time.NewTicker(time.Second / TickRate)
	defer ticker.Stop()

	lastUpdate := time.Now()

	for g.running {
		select {
		case <-g.stopChan:
			g.running = false
			return

		case input := <-g.InputChan:
			g.processInput(input)

		case <-ticker.C:
			now := time.Now()
			dt := now.Sub(lastUpdate).Seconds()
			lastUpdate = now

			g.update(dt)
			g.broadcast()
		}
	}
}

// Stop halts the game loop
func (g *Game) Stop() {
	if g.running {
		close(g.stopChan)
	}
}

// processInput handles a player's input
func (g *Game) processInput(input PlayerInput) {
	g.mu.Lock()
	defer g.mu.Unlock()

	tank, exists := g.Tanks[input.PlayerID]
	if !exists {
		return
	}

	switch input.Action {
	case "input":
		tank.Input = input.Input
	case "fire":
		if bullet := tank.Fire(); bullet != nil {
			g.Bullets = append(g.Bullets, bullet)
		}
	case "switch_weapon":
		tank.SwitchWeapon(input.Weapon)
	}
}

// update processes one frame of game logic
func (g *Game) update(dt float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.State != StatePlaying {
		return
	}

	bounds := g.Map.GetBounds()

	// Update tanks
	for _, tank := range g.Tanks {
		tank.Update(dt, bounds)

		// Auto-fire if holding mouse button
		if tank.Input.Firing {
			if bullet := tank.Fire(); bullet != nil {
				g.Bullets = append(g.Bullets, bullet)
			}
		}
	}

	// Update bullets
	activeBullets := make([]*Bullet, 0)
	for _, bullet := range g.Bullets {
		if bullet.Update(dt) {
			// Check if bullet is still in bounds
			if bounds.Contains(bullet.Position) || bullet.Type == BulletMortar {
				activeBullets = append(activeBullets, bullet)
			}
		}
	}
	g.Bullets = activeBullets

	// Check collisions
	g.checkCollisions()
}

// checkCollisions handles bullet-tank collisions
func (g *Game) checkCollisions() {
	bulletsToRemove := make(map[int]bool)

	for i, bullet := range g.Bullets {
		if !bullet.IsActive() {
			continue
		}

		bulletHitbox := bullet.GetHitbox()

		for _, tank := range g.Tanks {
			// Don't hit yourself with normal bullets
			if bullet.OwnerID == tank.ID && bullet.Type == BulletNormal {
				continue
			}

			tankHitbox := tank.GetHitbox()

			if bulletHitbox.Intersects(tankHitbox) {
				tank.TakeDamage(bullet.Damage)
				bulletsToRemove[i] = true

				// Check for game over
				if !tank.IsAlive() {
					g.State = StateGameOver
					// Find the winner
					for id := range g.Tanks {
						if id != tank.ID {
							g.WinnerID = id
							break
						}
					}
				}
				break
			}
		}
	}

	// Remove hit bullets
	if len(bulletsToRemove) > 0 {
		newBullets := make([]*Bullet, 0)
		for i, bullet := range g.Bullets {
			if !bulletsToRemove[i] {
				newBullets = append(newBullets, bullet)
			}
		}
		g.Bullets = newBullets
	}
}

// broadcast sends the current game state to all players
func (g *Game) broadcast() {
	if g.BroadcastFn == nil {
		return
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	state := g.GetState()
	g.BroadcastFn(state)
}

// GameStateMessage is the state sent to clients
type GameStateMessage struct {
	Type     string                 `json:"type"`
	State    GameState              `json:"state"`
	Tanks    map[string]TankState   `json:"tanks"`
	Bullets  []BulletState          `json:"bullets"`
	Map      MapConfig              `json:"map"`
	WinnerID string                 `json:"winnerId,omitempty"`
}

// GetState returns the current game state for broadcasting
func (g *Game) GetState() GameStateMessage {
	tanks := make(map[string]TankState)
	for id, tank := range g.Tanks {
		tanks[id] = tank.ToState()
	}

	bullets := make([]BulletState, len(g.Bullets))
	for i, bullet := range g.Bullets {
		bullets[i] = bullet.ToState()
	}

	return GameStateMessage{
		Type:     "game_state",
		State:    g.State,
		Tanks:    tanks,
		Bullets:  bullets,
		Map:      g.Map,
		WinnerID: g.WinnerID,
	}
}

// HandleInput queues input for processing
func (g *Game) HandleInput(input PlayerInput) {
	select {
	case g.InputChan <- input:
	default:
		// Channel full, drop input
	}
}
