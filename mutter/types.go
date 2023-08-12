package mutter

// https://gitlab.gnome.org/GNOME/mutter/-/blob/main/data/dbus-interfaces/org.gnome.Mutter.DisplayConfig.xml

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
}

type applyMethod uint32

const (
	applyVerify     applyMethod = 0
	applyTemporary  applyMethod = 1
	applyPersistent applyMethod = 2
)
