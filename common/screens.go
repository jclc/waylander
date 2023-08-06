package common

import "math"

const epsilon = 0.001 // for comparing floats

type Mode struct {
	Dimensions Rect    `json:"dimensions"`
	Frequency  float64 `json:"frequency"`
}

func ModesEqual(a, b Mode) bool {
	return a.Dimensions.Eq(b.Dimensions) &&
		math.Abs(a.Frequency-b.Frequency) < epsilon
}

type Screen struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Serial  string `json:"serial"`
	Modes   []Mode `json:"-"`
}

type Output struct {
	Connector string
	Screen    Screen
}

// ScreenState represents one logical monitor. It can have one or more physical
// monitors as its outputs, in which case the same logical monitor is cloned on
// all the outputs.
type ScreenState struct {
	Outputs     map[string]Mode `json:"outputs"`
	Scale       float64         `json:"scale"`
	Orientation Orientation     `json:"orientation"`
	Offset      Rect            `json:"offset"`
	Primary     bool            `json:"primary"`
}

// Profile represents a complete monitor layout.
type Profile struct {
	Monitors []ScreenState `json:"monitors"`
}
