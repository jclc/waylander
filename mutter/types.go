package mutter

// https://gitlab.gnome.org/GNOME/mutter/-/blob/main/data/dbus-interfaces/org.gnome.Mutter.DisplayConfig.xml

type crtc struct {
	ID               uint32         `dbus:"ID"`
	WinsysID         int64          `dbus:"winsys_id"`
	X                int32          `dbus:"x"`
	Y                int32          `dbus:"y"`
	Width            int32          `dbus:"width"`
	Height           int32          `dbus:"height"`
	CurrentMode      int32          `dbus:"current_mode"`
	CurrentTransform uint32         `dbus:"current_transform"`
	Transforms       []uint32       `dbus:"transforms"`
	Properties       map[string]any `dbus:"properties"`
}

type output struct {
	ID            uint32         `dbus:"ID"`
	WinsysID      int64          `dbus:"winsys_id"`
	CurrentCRTC   int32          `dbus:"current_crtc"`
	PossibleCRTCs []uint32       `dbus:"possible_crtcs"`
	Name          string         `dbus:"name"`
	Modes         []uint32       `dbus:"modes"`
	Clones        []uint32       `dbus:"clones"`
	Properties    map[string]any `dbus:"properties"`
}

type resMode struct {
	ID        uint32  `dbus:"ID"`
	WinsysID  int64   `dbus:"winsys_id"`
	Width     uint32  `dbus:"width"`
	Height    uint32  `dbus:"height"`
	Frequency float64 `dbus:"frequency"`
	Flags     uint32  `dbus:"flags"`
}

type resources struct {
	CRTCs           []crtc    `dbus:"crtcs"`
	Outputs         []output  `dbus:"outputs"`
	Modes           []resMode `dbus:"modes"`
	MaxScreenWidth  uint32    `dbus:"max_screen_width"`
	MaxScreenHeight uint32    `dbus:"max_screen_height"`
}

type stMode struct {
	ID              string         `dbus:"id"`
	Width           int32          `dbus:"width"`
	Height          int32          `dbus:"height"`
	RefreshRate     float64        `dbus:"refresh_rate"`
	PreferredScale  float64        `dbus:"preferred_scale"`
	SupportedScales []float64      `dbus:"supported_scales"`
	Properties      map[string]any `dbus:"properties"`
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
	X          int32          `dbus:"x"`
	Y          int32          `dbus:"y"`
	Scale      float64        `dbus:"scale"`
	Transform  uint32         `dbus:"transform"`
	Primary    bool           `dbus:"primary"`
	Monitors   []monitorInfo  `dbus:"monitors"`
	Properties map[string]any `dbus:"properties"`
}

type state struct {
	Monitors        []monitor
	LogicalMonitors []logicalMonitor
	Properties      map[string]any
}
