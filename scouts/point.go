package scouts

import (
	"encoding"
	"fmt"
	"image"
)

// Point is a point on the board.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var (
	_ fmt.Stringer             = Point{}
	_ encoding.TextMarshaler   = Point{}
	_ encoding.TextUnmarshaler = (*Point)(nil)
)

// Pt returns a Point with the given coordinates.
func Pt(x, y int) Point {
	return Point{x, y}
}

// String returns a string representation of the point in the form "x,y".
func (p Point) String() string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (p Point) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (p *Point) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%d,%d", &p.X, &p.Y)
	return err
}

// Add returns the vector p+q.
func (p Point) Add(q Point) Point {
	return Point{p.X + q.X, p.Y + q.Y}
}

// Sub returns the vector p-q.
func (p Point) Sub(q Point) Point {
	return Point{p.X - q.X, p.Y - q.Y}
}

// Mul returns the vector p*k.
func (p Point) Mul(k int) Point {
	return Point{p.X * k, p.Y * k}
}

// Div returns the vector p/k.
func (p Point) Div(k int) Point {
	return Point{p.X / k, p.Y / k}
}

// In reports whether p is in r.
func (p Point) In(r image.Rectangle) bool {
	return r.Min.X <= p.X && p.X < r.Max.X &&
		r.Min.Y <= p.Y && p.Y < r.Max.Y
}

// Eq reports whether p and q are equal.
func (p Point) Eq(q Point) bool {
	return p == q
}
