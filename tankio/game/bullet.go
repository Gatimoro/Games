package game

import (
	"time"
)

// BulletType identifies the type of projectile
type BulletType string

const (
	BulletNormal BulletType = "normal"
	BulletMortar BulletType = "mortar"
)

// Bullet represents a projectile in the game
type Bullet struct {
	ID        string     `json:"id"`
	OwnerID   string     `json:"ownerId"`
	Position  Vector2    `json:"position"`
	Velocity  Vector2    `json:"velocity"`
	Type      BulletType `json:"type"`
	Damage    int        `json:"damage"`
	CreatedAt time.Time  `json:"-"`
	MaxAge    time.Duration `json:"-"`

	// Mortar-specific fields
	ImpactPos    Vector2   `json:"impactPos,omitempty"`
	ImpactTime   time.Time `json:"impactTime,omitempty"`
	ImpactRadius float64   `json:"impactRadius,omitempty"`
	HasImpacted  bool      `json:"-"`
}

// Update moves the bullet and returns false if it should be removed
func (b *Bullet) Update(dt float64) bool {
	// Check if bullet is too old
	if time.Since(b.CreatedAt) > b.MaxAge {
		return false
	}

	switch b.Type {
	case BulletNormal:
		// Normal bullets move linearly
		b.Position = b.Position.Add(b.Velocity.Scale(dt))
	case BulletMortar:
		// Mortar shells don't move - they have a fixed impact position
		// The visual arc is handled client-side
		if time.Now().After(b.ImpactTime) && !b.HasImpacted {
			b.HasImpacted = true
			b.Position = b.ImpactPos // Snap to impact position
		}
	}

	return true
}

// GetHitbox returns the collision circle for the bullet
func (b *Bullet) GetHitbox() Circle {
	radius := 5.0
	if b.Type == BulletMortar && b.HasImpacted {
		radius = b.ImpactRadius // Explosion radius
	}
	return Circle{Center: b.Position, Radius: radius}
}

// IsActive returns true if the bullet can deal damage
func (b *Bullet) IsActive() bool {
	switch b.Type {
	case BulletNormal:
		return true
	case BulletMortar:
		return b.HasImpacted // Only deals damage on impact
	}
	return false
}

// GetFlightProgress returns 0-1 for mortar shells (for client-side arc animation)
func (b *Bullet) GetFlightProgress() float64 {
	if b.Type != BulletMortar {
		return 1.0
	}
	totalFlight := b.ImpactTime.Sub(b.CreatedAt)
	elapsed := time.Since(b.CreatedAt)
	progress := float64(elapsed) / float64(totalFlight)
	if progress > 1.0 {
		return 1.0
	}
	return progress
}

// BulletState is the JSON-serializable state sent to clients
type BulletState struct {
	ID             string     `json:"id"`
	OwnerID        string     `json:"ownerId"`
	Position       Vector2    `json:"position"`
	Type           BulletType `json:"type"`
	ImpactPos      Vector2    `json:"impactPos,omitempty"`
	FlightProgress float64    `json:"flightProgress,omitempty"`
	ImpactRadius   float64    `json:"impactRadius,omitempty"`
}

// ToState converts a bullet to its client-visible state
func (b *Bullet) ToState() BulletState {
	return BulletState{
		ID:             b.ID,
		OwnerID:        b.OwnerID,
		Position:       b.Position,
		Type:           b.Type,
		ImpactPos:      b.ImpactPos,
		FlightProgress: b.GetFlightProgress(),
		ImpactRadius:   b.ImpactRadius,
	}
}
