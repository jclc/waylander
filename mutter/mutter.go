package mutter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/godbus/dbus/v5"
	"github.com/jclc/waylander/common"
)

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
		// Find preferred mode
	preferredModeLoop:
		for i := range s.st.Monitors {
			if s.st.Monitors[i].Info.Connector != o.Name {
				continue
			}
			for _, mode := range s.st.Monitors[i].Modes {
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

	physicalMonitors := map[string]monitor{}
	currentModes := map[string]stMode{}
	for _, s := range s.st.Monitors {
		physicalMonitors[s.Info.Connector] = s
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
	err := s.applyConfiguration(persistent, nil, nil)
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

func (s *session) applyConfiguration(persistent bool, crtcs []crtc, outputs []crtc) error {
	obj := s.conn.Object(
		"org.gnome.Mutter.DisplayConfig",
		"/org/gnome/Mutter/DisplayConfig")

	err := obj.Call("org.gnome.Mutter.DisplayConfig.ApplyConfiguration", 0,
		s.serial, persistent, crtcs, outputs).Err
	if err != nil {
		return fmt.Errorf("failed to call Mutter d-bus API: %w", err)
	}

	return nil
}
