package mutter

import "fmt"

// https://gitlab.gnome.org/GNOME/mutter/-/blob/main/data/dbus-interfaces/org.gnome.Mutter.DisplayConfig.xml

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
