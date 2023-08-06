package common

import (
	"fmt"
	"io"
)

type Orientation uint8

const (
	OrientNormal     Orientation = 0
	Orient90         Orientation = 1
	Orient180        Orientation = 2
	Orient270        Orientation = 3
	OrientFlipped    Orientation = 4
	Orient90Flipped  Orientation = 5
	Orient180Flipped Orientation = 6
	Orient270Flipped Orientation = 7
)

func (o Orientation) String() string {
	switch o {
	case OrientNormal:
		return "normal"
	case Orient90:
		return "90"
	case Orient180:
		return "180"
	case Orient270:
		return "270"
	case OrientFlipped:
		return "flipped"
	case Orient90Flipped:
		return "flipped90"
	case Orient180Flipped:
		return "flipped180"
	case Orient270Flipped:
		return "flipped270"
	}
	return "Unknown"
}

func (o Orientation) MarshalText() ([]byte, error) {
	return []byte(o.String()), nil
}

func (o *Orientation) UnmarshalText(text []byte) error {
	switch string(text) {
	case "normal":
		*o = OrientNormal
	case "90":
		*o = Orient90
	case "180":
		*o = Orient180
	case "270":
		*o = Orient270
	case "flipped":
		*o = OrientFlipped
	case "flipped90":
		*o = Orient90Flipped
	case "flipped180":
		*o = Orient180Flipped
	case "flipped270":
		*o = Orient270Flipped
	default:
		return fmt.Errorf("invalid orientation: %s", string(text))
	}
	return nil
}

type DesktopSession interface {
	Resources() (Resources, error)
	ScreenStates() ([]LogicalMonitor, error)
	Apply(profile Profile, persistent bool) error
	Close()
	DebugInfo(output io.Writer) error
}

// State is the output of the state command
type State struct {
	Monitors []LogicalMonitor `json:"monitors"`
}

// Resources is the output of the resources command
type Resources struct {
	// Monitors maps connector to connected physical monitors
	Monitors map[string]PhysicalMonitor `json:"monitors"`
}
