package mutter

// https://gitlab.gnome.org/GNOME/mutter/-/blob/main/data/dbus-interfaces/org.gnome.Mutter.DisplayConfig.xml

type crtc struct {
	ID               uint32
	WinsysID         int64
	X                int32
	Y                int32
	Width            int32
	Height           int32
	CurrentMode      int32
	CurrentTransform uint32
	Transforms       []uint32
	Properties       map[string]any
}

type output struct {
	ID            uint32
	WinsysID      int64
	CurrentCRTC   int32
	PossibleCRTCs []uint32
	Name          string
	Modes         []uint32
	Clones        []uint32
	Properties    map[string]any
}

type resMode struct {
	ID        uint32
	WinsysID  int64
	Width     uint32
	Height    uint32
	Frequency float64
	Flags     uint32
}

type resources struct {
	CRTCs           []crtc
	Outputs         []output
	Modes           []resMode
	MaxScreenWidth  uint32
	MaxScreenHeight uint32
}

type stMode struct {
	ID              string
	Width           int32
	Height          int32
	RefreshRate     float64
	PreferredScale  float64
	SupportedScales []float64
	Properties      map[string]any
}

type monitorInfo struct {
	Connector string
	Vendor    string
	Product   string
	Serial    string
}

type monitor struct {
	Info       monitorInfo
	Modes      []stMode
	Properties map[string]any
}

// type applyOutput struct {
// 	ID         uint32
// 	Properties map[string]any
// }

// type applyCrtc struct {
// 	ID         uint32
// 	NewMode    int32
// 	X          int32
// 	Y          int32
// 	Transform  uint32
// 	Outputs    []uint32
// 	Properties map[string]any
// }

type logicalMonitor struct {
	X          int32
	Y          int32
	Scale      float64
	Transform  uint32
	Primary    bool
	Monitors   []monitorInfo
	Properties map[string]any
}

type state struct {
	Monitors        []monitor
	LogicalMonitors []logicalMonitor
	Properties      map[string]any
}

type applyMonitor struct {
	Connector  string
	ModeID     string
	Properties map[string]any
}

type applyLogicalMonitor struct {
	X         int32
	Y         int32
	Scale     float64
	Transform uint32
	Primary   bool
	Monitors  []applyMonitor
	// Properties map[string]any
}

type applyMethod uint32

const (
	applyVerify     applyMethod = 0
	applyTemporary  applyMethod = 1
	applyPersistent applyMethod = 2
)
