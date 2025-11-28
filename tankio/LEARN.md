# Tank.io - Learning Go Through Game Development

This document explains the codebase in detail, highlighting Go idioms, patterns, and quirks you'll encounter.

## Table of Contents
1. [Quick Start](#quick-start)
2. [Project Structure](#project-structure)
3. [Go Fundamentals Used](#go-fundamentals-used)
4. [File-by-File Breakdown](#file-by-file-breakdown)
5. [Go Quirks & Gotchas](#go-quirks--gotchas)
6. [How WebSockets Work Here](#how-websockets-work-here)
7. [The Game Loop Pattern](#the-game-loop-pattern)

---

## Quick Start

```bash
# On Arch Linux, install Go if needed:
sudo pacman -S go

# Run the game:
cd tankio
go run .

# Or build and run:
go build -o tankio .
./tankio
```

Open `http://localhost:8080` in your browser.

---

## Project Structure

```
tankio/
├── main.go              # Entry point - sets up HTTP server
├── go.mod               # Module definition (like package.json)
├── go.sum               # Dependency checksums (like package-lock.json)
│
├── game/                # Core game logic (no network code)
│   ├── vector.go        # 2D math utilities
│   ├── tank.go          # Tank entity
│   ├── bullet.go        # Projectile entity
│   ├── weapon.go        # Weapon interface + implementations
│   ├── map.go           # Map configuration
│   └── game.go          # Game loop, state management
│
├── network/             # Network layer
│   ├── messages.go      # Message type definitions
│   ├── client.go        # Per-player WebSocket handler
│   ├── lobby.go         # Game room management
│   └── manager.go       # Lobby lifecycle, HTTP handlers
│
└── static/              # Client-side files
    ├── index.html       # UI structure + CSS
    └── game.js          # Canvas rendering, input handling
```

---

## Go Fundamentals Used

### 1. Packages and Imports

Every Go file starts with a package declaration:

```go
package game  // This file belongs to the "game" package
```

Import paths are based on the module name in `go.mod`:

```go
import (
    "tankio/game"      // Our game package
    "tankio/network"   // Our network package
    "time"             // Standard library
    "github.com/gorilla/websocket"  // External dependency
)
```

**Go quirk**: Unused imports cause compilation errors. This is intentional - Go enforces clean code.

### 2. Structs (Go's "Classes")

Go doesn't have classes. Instead, you use **structs** with **methods**:

```go
// Define a struct (like a class without methods)
type Tank struct {
    ID       string
    Position Vector2
    Health   int
}

// Attach a method to the struct
// (t *Tank) is called the "receiver" - like "this" in other languages
func (t *Tank) TakeDamage(amount int) {
    t.Health -= amount
}

// Usage:
tank := &Tank{ID: "player1", Health: 100}
tank.TakeDamage(25)  // tank.Health is now 75
```

**Pointer receivers (`*Tank`)** vs **Value receivers (`Tank`)**:
- Use `*Tank` when you need to modify the struct
- Use `Tank` when you only read from it (makes a copy)

### 3. Interfaces (Go's Polymorphism)

Interfaces define behavior, not data:

```go
// weapon.go
type Weapon interface {
    Fire(origin Vector2, angle float64, ownerID string) *Bullet
    CanFire() bool
    GetType() WeaponType
}
```

**Go quirk**: Interfaces are implemented *implicitly*. You don't write `implements Weapon`:

```go
// Cannon implements Weapon because it has all the methods
type Cannon struct {
    lastFired time.Time
    cooldown  time.Duration
}

func (c *Cannon) Fire(origin Vector2, angle float64, ownerID string) *Bullet { ... }
func (c *Cannon) CanFire() bool { ... }
func (c *Cannon) GetType() WeaponType { ... }
// Now Cannon automatically satisfies the Weapon interface!
```

### 4. Goroutines (Lightweight Threads)

Start a goroutine with the `go` keyword:

```go
// main.go
go lobbyManager.CleanupRoutine()  // Runs in background

// game.go
go g.Run()  // Start game loop in background
```

**Key difference from threads**: Goroutines are managed by Go's runtime, not the OS. You can have thousands of them efficiently.

### 5. Channels (Goroutine Communication)

Channels are Go's way of safely passing data between goroutines:

```go
// Create a channel that carries PlayerInput values
InputChan chan PlayerInput

// Send to channel (blocks if buffer full)
g.InputChan <- input

// Receive from channel (blocks if empty)
input := <-g.InputChan
```

**Buffered vs Unbuffered**:
```go
make(chan int)       // Unbuffered - send blocks until receive
make(chan int, 100)  // Buffered - can hold 100 items before blocking
```

### 6. Select Statement (Channel Multiplexing)

`select` lets you wait on multiple channels simultaneously:

```go
// game.go - The game loop
for g.running {
    select {
    case <-g.stopChan:           // Stop signal received
        return
    case input := <-g.InputChan: // Player input received
        g.processInput(input)
    case <-ticker.C:             // Tick timer fired
        g.update(dt)
        g.broadcast()
    }
}
```

**Go quirk**: If multiple cases are ready, one is chosen randomly.

### 7. Defer (Cleanup)

`defer` schedules a function to run when the current function returns:

```go
func (g *Game) update(dt float64) {
    g.mu.Lock()
    defer g.mu.Unlock()  // Will unlock even if we return early or panic

    // ... rest of function
}
```

**Execution order**: Defers run in LIFO order (last defer runs first).

### 8. JSON Tags (Serialization)

Struct tags control JSON encoding:

```go
type Tank struct {
    ID       string  `json:"id"`           // JSON key is "id"
    Position Vector2 `json:"position"`     // JSON key is "position"
    Speed    float64 `json:"-"`            // Excluded from JSON
    WinnerID string  `json:"winnerId,omitempty"` // Omit if empty
}
```

---

## File-by-File Breakdown

### `main.go` - Entry Point

```go
func main() {
    lobbyManager := network.NewLobbyManager()
    go lobbyManager.CleanupRoutine()  // Background cleanup

    // Route setup
    http.Handle("/", http.FileServer(http.Dir("static")))
    http.HandleFunc("/api/create-lobby", lobbyManager.HandleCreateLobby)
    http.HandleFunc("/ws", lobbyManager.HandleWebSocket)

    http.ListenAndServe(":8080", nil)
}
```

**Pattern**: Dependency injection - `LobbyManager` is created once and its methods handle routes.

---

### `game/vector.go` - Math Utilities

Value types for 2D math:

```go
type Vector2 struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

// Methods return new vectors (immutable pattern)
func (v Vector2) Add(other Vector2) Vector2 {
    return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}
```

**Go quirk**: These use value receivers (`Vector2` not `*Vector2`) because:
1. Vector2 is small (16 bytes)
2. We want immutable behavior (returns new vector)

---

### `game/weapon.go` - Interface Pattern

Demonstrates Go's interface pattern:

```go
// Interface definition
type Weapon interface {
    Fire(origin Vector2, angle float64, ownerID string) *Bullet
    CanFire() bool
    GetType() WeaponType
    GetAmmo() int
    GetMaxAmmo() int
    Update(dt float64)
}

// Two implementations: Cannon and Mortar
type Cannon struct { ... }
type Mortar struct { ... }
```

**Thread-safe ID generation**:
```go
import "sync/atomic"

var idCounter int64

func generateID() string {
    id := atomic.AddInt64(&idCounter, 1)  // Atomic increment
    return fmt.Sprintf("B%d", id)
}
```

---

### `game/tank.go` - Entity with Composition

Tank "has-a" weapons instead of "is-a" relationship:

```go
type Tank struct {
    Position     Vector2
    Cannon       *Cannon   // Composition: Tank HAS a Cannon
    Mortar       *Mortar   // Composition: Tank HAS a Mortar
    ActiveWeapon WeaponType
}

func (t *Tank) Fire() *Bullet {
    switch t.ActiveWeapon {
    case WeaponCannon:
        return t.Cannon.Fire(t.Position, t.TurretAngle, t.ID)
    case WeaponMortar:
        return t.Mortar.Fire(t.Position, t.TurretAngle, t.ID)
    }
    return nil
}
```

---

### `game/game.go` - Game Loop & Concurrency

The heart of the game - shows several Go patterns:

**Mutex for thread safety**:
```go
type Game struct {
    mu sync.RWMutex  // Protects shared state
    Tanks map[string]*Tank
    // ...
}

func (g *Game) update(dt float64) {
    g.mu.Lock()         // Exclusive access for writing
    defer g.mu.Unlock()
    // ... modify state safely
}

func (g *Game) GetState() GameStateMessage {
    g.mu.RLock()        // Shared access for reading
    defer g.mu.RUnlock()
    // ... read state safely
}
```

**Channel-based game loop**:
```go
func (g *Game) Run() {
    ticker := time.NewTicker(time.Second / 60)  // 60 FPS
    defer ticker.Stop()

    for g.running {
        select {
        case <-g.stopChan:
            return
        case input := <-g.InputChan:
            g.processInput(input)
        case <-ticker.C:
            g.update(dt)
            g.broadcast()
        }
    }
}
```

**Non-blocking channel send**:
```go
func (g *Game) HandleInput(input PlayerInput) {
    select {
    case g.InputChan <- input:  // Try to send
    default:                     // Channel full, drop input
    }
}
```

---

### `network/client.go` - WebSocket Pumps

Classic Go WebSocket pattern with two goroutines per connection:

```go
// ReadPump: WebSocket -> Server
func (c *Client) ReadPump() {
    defer func() {
        c.Lobby.Unregister <- c  // Cleanup on exit
        c.Conn.Close()
    }()

    for {
        _, message, err := c.Conn.ReadMessage()  // Blocks
        if err != nil {
            break
        }
        c.handleMessage(message)
    }
}

// WritePump: Server -> WebSocket
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer ticker.Stop()

    for {
        select {
        case message := <-c.Send:    // Message to send
            c.Conn.WriteMessage(websocket.TextMessage, message)
        case <-ticker.C:              // Ping timer
            c.Conn.WriteMessage(websocket.PingMessage, nil)
        case <-c.done:                // Shutdown signal
            return
        }
    }
}
```

**Why two goroutines?** WebSocket read/write can block. Separating them allows:
- Reading client input while sending game state
- Sending pings while waiting for messages

---

### `network/lobby.go` - Channel-Based Hub

Manages clients using channels (no locks needed for client map):

```go
type Lobby struct {
    Clients    map[string]*Client
    Register   chan *Client   // Add client
    Unregister chan *Client   // Remove client
    Broadcast  chan []byte    // Send to all
}

func (l *Lobby) Run() {
    for {
        select {
        case client := <-l.Register:
            l.Clients[client.ID] = client
        case client := <-l.Unregister:
            delete(l.Clients, client.ID)
        case message := <-l.Broadcast:
            for _, client := range l.Clients {
                client.Send <- message
            }
        }
    }
}
```

**Pattern**: Single goroutine owns the map, others communicate via channels. No mutex needed!

---

### `network/manager.go` - HTTP Handlers

Shows Go's HTTP handling:

```go
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
    json.NewEncoder(w).Encode(map[string]string{"code": lobby.Code})
}
```

---

## Go Quirks & Gotchas

### 1. Zero Values
Variables are initialized to their zero value:
```go
var i int      // 0
var s string   // ""
var p *Tank    // nil
var m map[string]int  // nil (must use make()!)
```

**Gotcha**: Nil maps panic on write:
```go
var m map[string]int
m["key"] = 1  // PANIC!

// Correct:
m := make(map[string]int)
m["key"] = 1  // OK
```

### 2. Slices vs Arrays
```go
arr := [3]int{1, 2, 3}     // Array - fixed size
slice := []int{1, 2, 3}    // Slice - dynamic size
slice = append(slice, 4)   // Grows automatically
```

### 3. Error Handling
Go doesn't have exceptions. Functions return errors:
```go
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
// Use conn...
```

### 4. Capitalization = Visibility
```go
type Tank struct {
    ID       string  // Exported (public) - uppercase
    position Vector2 // Unexported (private) - lowercase
}
```

### 5. Range Loop Gotcha
```go
tanks := []*Tank{tank1, tank2}
for _, tank := range tanks {
    go func() {
        fmt.Println(tank.ID)  // BUG: Always prints last tank!
    }()
}

// Fix: Copy the variable
for _, tank := range tanks {
    tank := tank  // Shadow the variable
    go func() {
        fmt.Println(tank.ID)  // Correct!
    }()
}
```

### 6. Interface Nil Gotcha
```go
var w Weapon = nil
if w == nil { /* true */ }

var c *Cannon = nil
var w Weapon = c
if w == nil { /* FALSE! Interface holds typed nil */ }
```

---

## How WebSockets Work Here

Unlike Arduino polling, WebSockets are **event-driven** and **bidirectional**:

```
1. Client connects:
   Browser ──── HTTP GET /ws?lobby=ABCD ────> Server
   Browser <─── HTTP 101 Switching Protocols ── Server

2. Connection upgraded to WebSocket (persistent)
   Browser <════════════════════════════════> Server

3. Client sends only when needed:
   Browser ── {"type":"input","payload":{...}} ─> Server
   (Only on keypress/mouse move, not polling!)

4. Server pushes at fixed rate:
   Browser <── {"type":"game_state",...} ─────── Server
   (60 times per second)
```

**Key insight**: The client doesn't ask "what's the game state?" repeatedly. The server *pushes* state updates.

---

## The Game Loop Pattern

```
┌────────────────────────────────────────────────┐
│                 Game.Run()                      │
│                                                │
│   ┌──────────┐                                 │
│   │  select  │ ← Waits for any of:             │
│   └────┬─────┘                                 │
│        │                                       │
│   ┌────┴────┬──────────────┬────────────┐     │
│   ▼         ▼              ▼            ▼     │
│ stopChan  InputChan    ticker.C     (repeat)  │
│   │         │              │                   │
│   │    processInput()   update(dt)            │
│   │         │           broadcast()           │
│   │         │              │                   │
│   ▼         └──────────────┘                   │
│  return                                        │
└────────────────────────────────────────────────┘
```

The `select` statement is the heart - it multiplexes:
- Stop signals (shutdown gracefully)
- Player input (process immediately)
- Tick timer (update physics, send state)

This is more efficient than a traditional `while(true)` loop because `select` blocks until something happens - no busy-waiting.

---

## Next Steps

1. **Add obstacles**: Implement `Rock` and `Wall` in `map.go`
2. **Client prediction**: Reduce perceived lag by predicting movement locally
3. **Sound effects**: Use Web Audio API in `game.js`
4. **Visual polish**: Particle effects, screen shake
5. **Tests**: Add unit tests with `go test`

Happy coding!
