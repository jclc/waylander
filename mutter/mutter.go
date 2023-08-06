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

	if err = s.getResources(); err != nil {
		return nil, err
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

func (s *session) Outputs() []common.Output {
	var outputs []common.Output
	for _, output := range s.res.Outputs {
		modes := make([]common.Mode, 0, len(output.Modes))
		for _, i := range output.Modes {
			m := s.res.Modes[i]
			modes = append(modes, common.Mode{
				Dimensions: common.Rect{
					X: int(m.Width),
					Y: int(m.Height),
				},
				Frequency: m.Frequency,
			})
		}
		outputs = append(outputs, common.Output{
			Connector: output.Name,
			Screen: common.Screen{
				Product: output.Properties["product"].(string),
				Vendor:  output.Properties["vendor"].(string),
				Serial:  output.Properties["serial"].(string),
				Modes:   modes,
			},
		})
	}
	return outputs
}

func (s *session) ScreenStates() []common.ScreenState {
	err := s.getState()
	if err != nil {
		fmt.Println(err)
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

	var states []common.ScreenState
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

		states = append(states, common.ScreenState{
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
	return states
}

func (s *session) Apply(profile common.Profile, persistent bool) error {
	err := s.applyConfiguration(persistent, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *session) DebugInfo(output io.Writer) {
	err := s.getResources()
	if err != nil {
		panic(err)
	}
	err = s.getState()
	if err != nil {
		panic(err)
	}

	dg := struct {
		Resources resources
		State     state
	}{
		Resources: s.res,
		State:     s.st,
	}

	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	err = enc.Encode(dg)
	if err != nil {
		panic(err)
	}
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
