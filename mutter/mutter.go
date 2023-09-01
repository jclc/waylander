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

const (
	epsilon          = 0.005
	preferredString  = "is-preferred"
	currentString    = "is-current"
	vrrCapableString = "is-vrr-allowed"
	vrrEnabledString = "allow_vrr"
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
	st     state
}

func (s *session) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *session) Resources() (common.Resources, error) {
	if err := s.getState(); err != nil {
		return common.Resources{}, err
	}

	res := common.Resources{
		Monitors: map[string]common.PhysicalMonitor{},
	}
	for _, o := range s.st.Monitors {
		mon := common.PhysicalMonitor{
			Vendor:     o.Info.Vendor,
			Product:    o.Info.Product,
			Serial:     o.Info.Serial,
			Properties: map[string]any{},
		}

		mon.Modes = make([]common.Mode, 0, len(o.Modes))
		for _, mode := range o.Modes {
			newMode := common.Mode{
				Dimensions: common.Rect{
					X: int(mode.Width),
					Y: int(mode.Height),
				},
				Frequency: mode.RefreshRate,
			}
			mon.Modes = append(mon.Modes, newMode)

			if common.GetProperty[bool](mode.Properties, preferredString) {
				mon.PreferredMode = newMode
			}
		}

		mon.Properties[common.PropertyVRRSupported] = common.GetProperty[bool](
			o.Properties, vrrCapableString)

		res.Monitors[o.Info.Connector] = mon
	}
	return res, nil
}

func (s *session) ScreenStates() ([]common.LogicalMonitor, error) {
	err := s.getState()
	if err != nil {
		return nil, err
	}

	currentModes := map[string]stMode{}
	for _, s := range s.st.Monitors {
		for _, m := range s.Modes {
			if common.GetProperty[bool](m.Properties, currentString) {
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

func (s *session) Apply(profile common.Profile, verify, persistent bool) error {
	err := s.getState()
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
		var scale float64
		for _, connector := range connectors {
			var id string
			id, scale = s.findModeID(connector, mon.Outputs[connector], mon.Scale)

			monitors = append(monitors, applyMonitor{
				Connector: connector,
				ModeID:    id,
				Properties: map[string]any{
					vrrEnabledString: common.GetProperty[bool](
						mon.Properties, common.PropertyVRREnabled,
					),
				},
			})
		}
		outputMonitors = append(outputMonitors, applyLogicalMonitor{
			X:         int32(mon.Offset.X),
			Y:         int32(mon.Offset.Y),
			Scale:     scale,
			Transform: uint32(mon.Orientation),
			Primary:   mon.Primary,
			Monitors:  monitors,
		})
	}

	method := applyTemporary
	if verify {
		// this is a GNOME oddity
		method = applyPersistent
	}

	err = s.applyMonitorsConfig(method, outputMonitors, nil)
	if err != nil {
		return err
	}
	if persistent {
		err = s.applyMonitorsConfig(applyPersistent, outputMonitors, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *session) DebugInfo(output io.Writer) error {
	err := s.getState()
	if err != nil {
		return err
	}

	dg := struct {
		Session string
		Serial  uint32
		State   state
	}{
		Session: "gnome",
		Serial:  s.serial,
		State:   s.st,
	}

	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")
	err = enc.Encode(dg)
	if err != nil {
		return err
	}

	return nil
}

func (s *session) findModeID(connector string, mode common.Mode, scale float64) (string, float64) {
	var best stMode
	for _, monitor := range s.st.Monitors {
		if monitor.Info.Connector != connector {
			continue
		}

		for _, m := range monitor.Modes {
			if int(m.Width) == mode.Dimensions.X &&
				int(m.Height) == mode.Dimensions.Y {
				// Choose the mode whose refresh rate best matches
				// the requested frequency
				curDelta := math.Abs(mode.Frequency - best.RefreshRate)
				newDelta := math.Abs(mode.Frequency - m.RefreshRate)
				if newDelta < curDelta {
					best = m
				}
			}
		}
	}
	if math.Abs(best.RefreshRate-mode.Frequency) > common.MaxAllowedFrequencyDeviation {
		panic(fmt.Sprintf("no matching mode %s found for %s", mode, connector))
	}

	newScale := common.Closest(best.SupportedScales, scale)

	return best.ID, newScale
}
