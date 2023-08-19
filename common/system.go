package common

import (
	"errors"
	"os"
	"path/filepath"
)

const configSubDir = "waylander"

var configPath string

func init() {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configPath = filepath.Join(xdgConfig, configSubDir)
	} else if home := os.Getenv("HOME"); home != "" {
		configPath = filepath.Join(home, ".config", configSubDir)
	} else {
		panic(errors.New("couldn't determine config directory"))
	}
}

func EnsureConfigDir() {
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Join(configPath, "profiles"), 0755)
	if err != nil {
		panic(err)
	}
}

func GetConfigDir() string {
	return configPath
}
