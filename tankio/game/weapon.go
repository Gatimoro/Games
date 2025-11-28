package game

import (
	"time"
)

// WeaponType identifies the type of weapon
type WeaponType string

const (
	WeaponCannon WeaponType = "cannon"
	WeaponMortar WeaponType = "mortar"
)

// Weapon interface defines behavior for all weapons
type Weapon interface {
	Fire(origin Vector2, angle float64, ownerID string) *Bullet
	CanFire() bool
	GetType() WeaponType
	GetAmmo() int      // -1 for unlimited
	GetMaxAmmo() int   // -1 for unlimited
	Update(dt float64) // Update cooldowns, recharge, etc.
}

// Cannon is the primary weapon - instant hit, unlimited ammo
type Cannon struct {
	lastFired time.Time
	cooldown  time.Duration
}

// NewCannon creates a new cannon weapon
func NewCannon() *Cannon {
	return &Cannon{
		cooldown: 500 * time.Millisecond, // 0.5 second between shots
	}
}

func (c *Cannon) Fire(origin Vector2, angle float64, ownerID string) *Bullet {
	if !c.CanFire() {
		return nil
	}
	c.lastFired = time.Now()

	direction := FromAngle(angle)
	return &Bullet{
		ID:        generateID(),
		OwnerID:   ownerID,
		Position:  origin.Add(direction.Scale(30)), // Spawn slightly ahead of tank
		Velocity:  direction.Scale(600),            // 600 pixels per second
		Type:      BulletNormal,
		Damage:    100, // Instant kill
		CreatedAt: time.Now(),
		MaxAge:    3 * time.Second,
	}
}

func (c *Cannon) CanFire() bool {
	return time.Since(c.lastFired) >= c.cooldown
}

func (c *Cannon) GetType() WeaponType {
	return WeaponCannon
}

func (c *Cannon) GetAmmo() int {
	return -1 // Unlimited
}

func (c *Cannon) GetMaxAmmo() int {
	return -1
}

func (c *Cannon) Update(dt float64) {
	// Cannon has no special update logic
}

// Mortar is the secondary weapon - delayed impact, limited ammo that recharges
type Mortar struct {
	lastFired     time.Time
	cooldown      time.Duration     // Time between shots
	ammo          int               // Current charges
	maxAmmo       int               // Maximum charges
	rechargeTime  time.Duration     // Time to regain one charge
	rechargeTimer time.Duration     // Current recharge progress
}

// NewMortar creates a new mortar weapon
func NewMortar() *Mortar {
	return &Mortar{
		cooldown:     3 * time.Second,  // 3 seconds between shots
		ammo:         3,                // Start with 3 charges
		maxAmmo:      3,
		rechargeTime: 10 * time.Second, // 10 seconds to regain a charge
	}
}

func (m *Mortar) Fire(origin Vector2, angle float64, ownerID string) *Bullet {
	if !m.CanFire() {
		return nil
	}
	m.lastFired = time.Now()
	m.ammo--

	direction := FromAngle(angle)
	// Mortar lands at a fixed distance (could be cursor position in future)
	impactDistance := 300.0
	impactPos := origin.Add(direction.Scale(impactDistance))

	return &Bullet{
		ID:          generateID(),
		OwnerID:     ownerID,
		Position:    origin,
		Velocity:    Vector2{X: 0, Y: 0}, // Mortar doesn't move linearly
		Type:        BulletMortar,
		Damage:      100, // Instant kill
		CreatedAt:   time.Now(),
		MaxAge:      4 * time.Second,           // Max flight time
		ImpactPos:   impactPos,
		ImpactTime:  time.Now().Add(3 * time.Second), // 3 second delay
		ImpactRadius: 50,                        // Explosion radius
	}
}

func (m *Mortar) CanFire() bool {
	return m.ammo > 0 && time.Since(m.lastFired) >= m.cooldown
}

func (m *Mortar) GetType() WeaponType {
	return WeaponMortar
}

func (m *Mortar) GetAmmo() int {
	return m.ammo
}

func (m *Mortar) GetMaxAmmo() int {
	return m.maxAmmo
}

func (m *Mortar) Update(dt float64) {
	// Recharge ammo over time
	if m.ammo < m.maxAmmo {
		m.rechargeTimer += time.Duration(dt * float64(time.Second))
		if m.rechargeTimer >= m.rechargeTime {
			m.ammo++
			m.rechargeTimer = 0
		}
	}
}

// Simple ID generator (in production, use UUID)
var idCounter int64

func generateID() string {
	idCounter++
	return string(rune('A'+idCounter%26)) + string(rune('0'+idCounter%10))
}
