package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Game holds all your game state
type Game struct {
	// TODO: Add your game state here
	// Example: playerX, playerY float64
}

// New creates a new game instance
func New() *Game {
	return &Game{
		// TODO: Initialize your state here
	}
}

// Update is called every tick (60 times per second by default)
// This is where you handle input and update game logic
func (g *Game) Update() error {
	// TODO: Handle input and update positions here
	// Example:
	// if ebiten.IsKeyPressed(ebiten.KeyW) {
	//     g.playerY -= 2
	// }

	return nil
}

// Draw is called every frame to render the screen
func (g *Game) Draw(screen *ebiten.Image) {
	// TODO: Draw your game here
	// Example:
	// ebitenutil.DrawRect(screen, g.playerX, g.playerY, 40, 40, color.White)
}

// Layout returns the logical screen size
// Ebitengine will scale this to fit the window
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}
