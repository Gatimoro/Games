package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"tankio/game"
)

func main() {
	// Create your game instance
	g := game.New()

	// Configure the window
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Tank.io")

	// Run the game loop - Ebitengine handles the loop for you
	// It calls g.Update() and g.Draw() automatically
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
