package common

import (
	"fmt"
	"math"
)

const (
	MaxAllowedFrequencyDeviation = 0.01
	Epsilon                      = 0.005 // for comparing floats
)

type Mode struct {
	Dimensions Rect    `json:"dimensions"`
	Frequency  float64 `json:"frequency"`
}

func (m Mode) String() string {
	return fmt.Sprintf("%dx%d @%f",
		m.Dimensions.X, m.Dimensions.Y, m.Frequency)
}

func (m Mode) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Mode) UnmarshalText(text []byte) error {
	var n Mode
	_, err := fmt.Sscanf(string(text), "%dx%d @%f",
		&n.Dimensions.X, &n.Dimensions.Y, &n.Frequency)
	if err != nil {
		return fmt.Errorf("error parsing mode: %w", err)
	}
	*m = n
	return nil
}

func ModesEqual(a, b Mode) bool {
	return a.Dimensions.Eq(b.Dimensions) &&
		math.Abs(a.Frequency-b.Frequency) < Epsilon
}

// LogicalMonitor represents one logical monitor. It can have one or more
// physical monitors as its outputs, in which case the same logical monitor is
// cloned to all of the outputs.
type LogicalMonitor struct {
	Outputs     map[string]Mode `json:"outputs"`
	Scale       float64         `json:"scale"`
	Orientation Orientation     `json:"orientation"`
	Offset      Rect            `json:"offset"`
	Primary     bool            `json:"primary"`
	VRREnabled  bool            `json:"vrr_enabled"` // variable refresh rate
}

// PhysicalMonitor represents one connected physical monitor output.
type PhysicalMonitor struct {
	Vendor        string `json:"vendor"`
	Product       string `json:"product"`
	Serial        string `json:"serial"`
	PreferredMode Mode   `json:"preferred_mode"`
	Modes         []Mode `json:"modes"`
	VRRSupported  bool   `json:"vrr_supported"` // variable refresh rate
}

// Profile represents a complete monitor layout.
type Profile struct {
	Monitors []LogicalMonitor `json:"monitors"`
}
