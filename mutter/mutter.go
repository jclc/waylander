package mutter

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"slices"

	"github.com/godbus/dbus/v5"
	"github.com/jclc/waylander/common"
	"golang.org/x/exp/maps"
)

const epsilon = 0.005

func GetDesktopSession() (common.DesktopSession, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to d-bus: %w", err)
	}
	s := &session{
		conn: conn,
	}

	return s, nil
}

type session struct {
	conn   *dbus.Conn
	serial uint32
	res    resources
	st     state
}

func (s *session) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *session) Resources() (common.Resources, error) {
	if err := s.getResources(); err != nil {
		return common.Resources{}, err
	}
	if err := s.getState(); err != nil {
		return common.Resources{}, err
	}

	res := common.Resources{
		Monitors: map[string]common.PhysicalMonitor{},
	}
	for _, o := range s.res.Outputs {
		mon := common.PhysicalMonitor{
			Vendor:  o.Properties["vendor"].(string),
			Product: o.Properties["product"].(string),
			Serial:  o.Properties["serial"].(string),
		}

		mon.Modes = make([]common.Mode, 0, len(o.Modes))
		for _, i := range o.Modes {
			mon.Modes = append(mon.Modes, common.Mode{
				Dimensions: common.Rect{
					X: int(s.res.Modes[i].Width),
					Y: int(s.res.Modes[i].Height),
				},
				Frequency: s.res.Modes[i].Frequency,
			})
		}

	preferredModeLoop:
		for _, m := range s.st.Monitors {
			if m.Info.Connector != o.Name {
				continue
			}

			_, supportsVRR := m.Properties["is-vrr-allowed"]
			mon.VRRSupported = supportsVRR

			for _, mode := range m.Modes {
				if isPreferred, found := mode.Properties["is-preferred"]; found && isPreferred.(bool) {
					mon.PreferredMode = common.Mode{
						Dimensions: common.Rect{
							X: int(mode.Width),
							Y: int(mode.Height),
						},
						Frequency: mode.RefreshRate,
					}
					break preferredModeLoop
				}
			}
		}

		res.Monitors[o.Name] = mon
	}
	return res, nil
}

func (s *session) ScreenStates() ([]common.LogicalMonitor, error) {
	err := s.getState()
	if err != nil {
		return nil, err
	}

	// physicalMonitors := map[string]monitor{}
	currentModes := map[string]stMode{}
	for _, s := range s.st.Monitors {
		// physicalMonitors[s.Info.Connector] = s
		for _, m := range s.Modes {
			if isCurrent, found := m.Properties["is-current"]; found && isCurrent.(bool) {
				currentModes[s.Info.Connector] = m
				break
			}
		}
	}

	var states []common.LogicalMonitor
	for _, m := range s.st.LogicalMonitors {
		inputs := make(map[string]common.Mode, len(m.Monitors))
		for _, in := range m.Monitors {
			mode := currentModes[in.Connector]
			inputs[in.Connector] = common.Mode{
				Dimensions: common.Rect{
					X: int(mode.Width),
					Y: int(mode.Height),
				},
				Frequency: mode.RefreshRate,
			}
		}

		states = append(states, common.LogicalMonitor{
			Outputs: inputs,
			Offset: common.Rect{
				X: int(m.X),
				Y: int(m.Y),
			},
			Scale:       m.Scale,
			Orientation: common.Orientation(m.Transform),
			Primary:     m.Primary,
		})
	}

	return states, nil
}

func (s *session) Apply(profile common.Profile, persistent bool) error {
	err := s.getResources()
	if err != nil {
		return err
	}
	err = s.getState()
	if err != nil {
		return err
	}

	var outputMonitors []applyLogicalMonitor
	for i, mon := range profile.Monitors {
		connectors := maps.Keys(mon.Outputs)
		slices.Sort(connectors)

		if len(connectors) == 0 {
			return fmt.Errorf("monitor #%d has no outputs", i)
		} else if len(mon.Outputs) > 1 {
			// When mirroring, all outputs must have the same dimensions
			comp := mon.Outputs[connectors[0]]
			for _, mode := range mon.Outputs {
				if mode.Dimensions.X != comp.Dimensions.X ||
					mode.Dimensions.Y != comp.Dimensions.Y {
					return fmt.Errorf(
						"cannot mirror outputs with different dimensions "+
							"(%d,%d) and (%d,%d)",
						comp.Dimensions.X, comp.Dimensions.Y,
						mode.Dimensions.X, mode.Dimensions.Y)
				}
			}
		}

		// TODO: check for overlapping logical monitors

		var monitors []applyMonitor
		for _, connector := range connectors {
			monitors = append(monitors, applyMonitor{
				Connector: connector,
				ModeID:    s.findModeID(connector, mon.Outputs[connector]),
			})
		}
		outputMonitors = append(outputMonitors, applyLogicalMonitor{
			X:         int32(mon.Offset.X),
			Y:         int32(mon.Offset.Y),
			Scale:     mon.Scale,
			Transform: uint32(mon.Orientation),
			Primary:   mon.Primary,
			Monitors:  monitors,
		})
	}

	method := applyTemporary
	if persistent {
		method = applyVerify
	}

	err = s.applyMonitorsConfig(method, outputMonitors, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *session) DebugInfo(output io.Writer) error {
	err := s.getResources()
	if err != nil {
		return err
	}
	err = s.getState()
	if err != nil {
		return err
	}

	dg := struct {
		Session   string
		Serial    uint32
		Resources resources
		State     state
	}{
		Session:   "gnome",
		Serial:    s.serial,
		Resources: s.res,
		State:     s.st,
	}

	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	err = enc.Encode(dg)
	if err != nil {
		return err
	}

	return nil
}

func (s *session) findModeID(connector string, mode common.Mode) string {
	for _, monitor := range s.st.Monitors {
		if monitor.Info.Connector != connector {
			continue
		}

		for _, m := range monitor.Modes {
			if int(m.Width) == mode.Dimensions.X &&
				int(m.Height) == mode.Dimensions.Y &&
				math.Abs(mode.Frequency-m.RefreshRate) < epsilon {
				return m.ID
			}
		}
	}

	panic("mode not found for " + connector)
}

func (s *session) findOutputID(connector string) (uint32, error) {
	for _, output := range s.res.Outputs {
		if output.Name == connector {
			return output.ID, nil
		}
	}

	return 0, fmt.Errorf("no output %s", connector)
}

func (s *session) getResources() error {
	obj := s.conn.Object(
		"org.gnome.Mutter.DisplayConfig",
		"/org/gnome/Mutter/DisplayConfig")

	err := obj.Call("org.gnome.Mutter.DisplayConfig.GetResources", 0).Store(
		&s.serial, &s.res.CRTCs, &s.res.Outputs,
		&s.res.Modes, &s.res.MaxScreenWidth, &s.res.MaxScreenHeight)
	if err != nil {
		return fmt.Errorf("failed to call Mutter d-bus API: %w", err)
	}

	return nil
}

func (s *session) getState() error {
	obj := s.conn.Object(
		"org.gnome.Mutter.DisplayConfig",
		"/org/gnome/Mutter/DisplayConfig")

	err := obj.Call("org.gnome.Mutter.DisplayConfig.GetCurrentState", 0).Store(
		&s.serial, &s.st.Monitors, &s.st.LogicalMonitors, &s.st.Properties)
	if err != nil {
		return fmt.Errorf("failed to call Mutter d-bus API: %w", err)
	}

	return nil
}

func (s *session) applyMonitorsConfig(method applyMethod, logicalMonitors []applyLogicalMonitor, properties map[string]any) error {
	obj := s.conn.Object(
		"org.gnome.Mutter.DisplayConfig",
		"/org/gnome/Mutter/DisplayConfig")

	err := obj.Call("org.gnome.Mutter.DisplayConfig.ApplyMonitorsConfig", 0,
		s.serial, method, logicalMonitors, properties).Err
	if err != nil {
		return fmt.Errorf("failed to call Mutter d-bus API: %w", err)
	}

	return nil
}
