package game

// MapConfig holds the map dimensions and settings
type MapConfig struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// DefaultMap returns the default map configuration
func DefaultMap() MapConfig {
	return MapConfig{
		Width:  1200,
		Height: 800,
	}
}

// GetBounds returns the map as a Rectangle
func (m MapConfig) GetBounds() Rectangle {
	return Rectangle{
		X:      0,
		Y:      0,
		Width:  m.Width,
		Height: m.Height,
	}
}

// GetSpawnPoints returns spawn positions for players
func (m MapConfig) GetSpawnPoints() []Vector2 {
	// Spawn players on opposite sides of the map
	return []Vector2{
		{X: 100, Y: m.Height / 2},                // Left side
		{X: m.Width - 100, Y: m.Height / 2},      // Right side
	}
}

// Obstacle interface for future wall/rock implementations
type Obstacle interface {
	GetBounds() Rectangle
	BlocksBullets() bool
	BlocksTanks() bool
	GetType() string
}

// TODO: Implement Rock and Wall obstacles
// type Rock struct { ... }
// type Wall struct { ... }
