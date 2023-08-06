package common

import "fmt"

type Rect struct {
	X int
	Y int
}

func (r Rect) Add(s Rect) Rect {
	return Rect{
		X: r.X + s.X,
		Y: r.Y + s.Y,
	}
}

func (r Rect) Sub(s Rect) Rect {
	return Rect{
		X: r.X - s.X,
		Y: r.Y - s.Y,
	}
}

func (r Rect) Eq(s Rect) bool {
	return r.X == s.X && r.Y == s.Y
}

func (r Rect) String() string {
	return fmt.Sprintf("%dx%d", r.X, r.Y)
}

func (r Rect) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Rect) UnmarshalText(text []byte) error {
	var s Rect
	_, err := fmt.Sscanf(string(text), "%dx%d", &s.X, &s.Y)
	if err != nil {
		return fmt.Errorf("error parsing coordinate: %w", err)
	}
	*r = s
	return nil
}
