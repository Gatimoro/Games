package game
import (
    "math"
    "github.com/hajimehoshi/ebiten/v2"
)

var invSqrt2 = 1 / math.Sqrt(2) 

const MAP_WIDTH int = 20
const MAP_HEIGHT int = 20
type GameState string
const (
	Waiting  GameState = "waiting"
	Playing  GameState = "playing"
	GameOver GameState = "gameover"
)
// Game holds all your game state
type Game struct {
	players []Player 
	blocks []Block 
	bullets []Bullet
	state GameState
	mapWidth int
	mapHeight int
}
// New creates a new game instance
func New() *Game {
    return &Game{
        players:    []Player{},
        blocks:     []Block{},
        bullets:    []Bullet{},
        state:      Waiting,
        mapWidth:  MAP_WIDTH,
        mapHeight: MAP_HEIGHT,
    }
}

// Update is called every tick (60 times per second by default)
// This is where you handle input and update game logic
func (g *Game) Update() error {
	switch g.state{
	case Waiting:
		if len(g.players) >= 2{
			g.SetState(Playing)
		}
	case Playing:
        	g.MovePlayers()
	case GameOver:

	}
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
func (g *Game) SetState(s GameState){
	g.state = s
}
func (g *Game) MovePlayers(){
    for i := range g.players{
        p := &g.players[i]  
        
        dx := float64((p.keys & 8) >> 3) - float64((p.keys & 2) >> 1)  // D - A
        dy := float64((p.keys & 4) >> 2) - float64(p.keys & 1)         // S - W
        
        if dx != 0 && dy != 0 {
            dx *= invSqrt2
            dy *= invSqrt2
        }
        
        nx, ny := p.x + dx, p.y + dy
        if g.inbounds(nx, ny){
            p.x, p.y = nx, ny
        }
    }
}
func (g *Game) inbounds(x_pos, y_pos float64) bool{
	return x_pos >= 0 && x_pos < float64(g.mapWidth) && y_pos >= 0 && y_pos < float64(g.mapHeight)
}
type Player struct{
	alive bool
	x, y float64
	keys byte
	look byte
}






