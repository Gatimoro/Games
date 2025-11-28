package game

import (
	"math"
)

const (
	TankSpeed       = 200.0 // Pixels per second
	TankSize        = 40.0  // Tank width/height
	TankTurretLength = 30.0
)

// Tank represents a player's tank
type Tank struct {
	ID            string     `json:"id"`
	Position      Vector2    `json:"position"`
	Rotation      float64    `json:"rotation"`      // Body rotation (movement direction)
	TurretAngle   float64    `json:"turretAngle"`   // Turret rotation (aim direction)
	Health        int        `json:"health"`
	MaxHealth     int        `json:"maxHealth"`
	Speed         float64    `json:"-"`
	ActiveWeapon  WeaponType `json:"activeWeapon"`
	Cannon        *Cannon    `json:"-"`
	Mortar        *Mortar    `json:"-"`

	// Input state (received from client)
	Input InputState `json:"-"`
}

// InputState represents the current input from the player
type InputState struct {
	Up       bool    `json:"up"`
	Down     bool    `json:"down"`
	Left     bool    `json:"left"`
	Right    bool    `json:"right"`
	MouseX   float64 `json:"mouseX"`
	MouseY   float64 `json:"mouseY"`
	Firing   bool    `json:"firing"`
}

// NewTank creates a new tank at the given position
func NewTank(id string, pos Vector2) *Tank {
	return &Tank{
		ID:           id,
		Position:     pos,
		Rotation:     0,
		TurretAngle:  0,
		Health:       100,
		MaxHealth:    100,
		Speed:        TankSpeed,
		ActiveWeapon: WeaponCannon,
		Cannon:       NewCannon(),
		Mortar:       NewMortar(),
	}
}

// Update processes tank movement and state for one frame
func (t *Tank) Update(dt float64, mapBounds Rectangle) {
	// Calculate movement direction from input
	var moveDir Vector2
	if t.Input.Up {
		moveDir.Y -= 1
	}
	if t.Input.Down {
		moveDir.Y += 1
	}
	if t.Input.Left {
		moveDir.X -= 1
	}
	if t.Input.Right {
		moveDir.X += 1
	}

	// Apply movement
	if moveDir.X != 0 || moveDir.Y != 0 {
		moveDir = moveDir.Normalize()
		newPos := t.Position.Add(moveDir.Scale(t.Speed * dt))

		// Update body rotation to face movement direction
		t.Rotation = moveDir.Angle()

		// Clamp to map bounds
		halfSize := TankSize / 2
		newPos.X = math.Max(mapBounds.X+halfSize, math.Min(newPos.X, mapBounds.X+mapBounds.Width-halfSize))
		newPos.Y = math.Max(mapBounds.Y+halfSize, math.Min(newPos.Y, mapBounds.Y+mapBounds.Height-halfSize))

		t.Position = newPos
	}

	// Update turret angle to point at mouse
	mousePos := Vector2{X: t.Input.MouseX, Y: t.Input.MouseY}
	toMouse := mousePos.Sub(t.Position)
	t.TurretAngle = toMouse.Angle()

	// Update weapons
	t.Cannon.Update(dt)
	t.Mortar.Update(dt)
}

// Fire attempts to fire the current weapon
func (t *Tank) Fire() *Bullet {
	switch t.ActiveWeapon {
	case WeaponCannon:
		return t.Cannon.Fire(t.Position, t.TurretAngle, t.ID)
	case WeaponMortar:
		return t.Mortar.Fire(t.Position, t.TurretAngle, t.ID)
	}
	return nil
}

// SwitchWeapon changes the active weapon
func (t *Tank) SwitchWeapon(weapon WeaponType) {
	t.ActiveWeapon = weapon
}

// TakeDamage applies damage to the tank
func (t *Tank) TakeDamage(amount int) {
	t.Health -= amount
	if t.Health < 0 {
		t.Health = 0
	}
}

// IsAlive returns true if the tank has health remaining
func (t *Tank) IsAlive() bool {
	return t.Health > 0
}

// GetHitbox returns the collision circle for the tank
func (t *Tank) GetHitbox() Circle {
	return Circle{Center: t.Position, Radius: TankSize / 2}
}

// GetCurrentWeapon returns the active weapon
func (t *Tank) GetCurrentWeapon() Weapon {
	switch t.ActiveWeapon {
	case WeaponCannon:
		return t.Cannon
	case WeaponMortar:
		return t.Mortar
	}
	return t.Cannon
}

// TankState is the JSON-serializable state sent to clients
type TankState struct {
	ID           string     `json:"id"`
	Position     Vector2    `json:"position"`
	Rotation     float64    `json:"rotation"`
	TurretAngle  float64    `json:"turretAngle"`
	Health       int        `json:"health"`
	MaxHealth    int        `json:"maxHealth"`
	ActiveWeapon WeaponType `json:"activeWeapon"`
	MortarAmmo   int        `json:"mortarAmmo"`
	MortarMaxAmmo int       `json:"mortarMaxAmmo"`
}

// ToState converts a tank to its client-visible state
func (t *Tank) ToState() TankState {
	return TankState{
		ID:           t.ID,
		Position:     t.Position,
		Rotation:     t.Rotation,
		TurretAngle:  t.TurretAngle,
		Health:       t.Health,
		MaxHealth:    t.MaxHealth,
		ActiveWeapon: t.ActiveWeapon,
		MortarAmmo:   t.Mortar.GetAmmo(),
		MortarMaxAmmo: t.Mortar.GetMaxAmmo(),
	}
}
