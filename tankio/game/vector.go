package game

import "math"

// Vector2 represents a 2D point or direction
type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Add returns the sum of two vectors
func (v Vector2) Add(other Vector2) Vector2 {
	return Vector2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference of two vectors
func (v Vector2) Sub(other Vector2) Vector2 {
	return Vector2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Scale multiplies the vector by a scalar
func (v Vector2) Scale(s float64) Vector2 {
	return Vector2{X: v.X * s, Y: v.Y * s}
}

// Length returns the magnitude of the vector
func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a unit vector in the same direction
func (v Vector2) Normalize() Vector2 {
	l := v.Length()
	if l == 0 {
		return Vector2{X: 0, Y: 0}
	}
	return Vector2{X: v.X / l, Y: v.Y / l}
}

// Distance returns the distance between two points
func (v Vector2) Distance(other Vector2) float64 {
	return v.Sub(other).Length()
}

// Angle returns the angle of the vector in radians
func (v Vector2) Angle() float64 {
	return math.Atan2(v.Y, v.X)
}

// FromAngle creates a unit vector from an angle in radians
func FromAngle(angle float64) Vector2 {
	return Vector2{X: math.Cos(angle), Y: math.Sin(angle)}
}

// Rectangle represents an axis-aligned bounding box
type Rectangle struct {
	X, Y, Width, Height float64
}

// Contains checks if a point is inside the rectangle
func (r Rectangle) Contains(p Vector2) bool {
	return p.X >= r.X && p.X <= r.X+r.Width &&
		p.Y >= r.Y && p.Y <= r.Y+r.Height
}

// Intersects checks if two rectangles overlap
func (r Rectangle) Intersects(other Rectangle) bool {
	return r.X < other.X+other.Width &&
		r.X+r.Width > other.X &&
		r.Y < other.Y+other.Height &&
		r.Y+r.Height > other.Y
}

// Circle represents a circular shape for collision
type Circle struct {
	Center Vector2
	Radius float64
}

// Contains checks if a point is inside the circle
func (c Circle) Contains(p Vector2) bool {
	return c.Center.Distance(p) <= c.Radius
}

// Intersects checks if two circles overlap
func (c Circle) Intersects(other Circle) bool {
	return c.Center.Distance(other.Center) <= c.Radius+other.Radius
}
