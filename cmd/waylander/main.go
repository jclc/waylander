package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jclc/waylander/common"
	"github.com/jclc/waylander/mutter"
)

var (
	session common.DesktopSession
)

func Usage() {
	fmt.Println("Waylander -- a screen management tool.")
	fmt.Println()
	fmt.Printf(
		"Usage: %s <command> [args...]\n"+
			"\n"+
			"Commands:\n"+
			"    help        Print this help message\n"+
			"    resources   Show currently connected outputs\n"+
			"    state       Show the current configuration\n"+
			"    profiles    List saved profiles\n"+
			"    save        Save current state as a profile\n"+
			"    apply       Apply a profile\n"+
			"    edit        Edit profile\n"+
			"    debuginfo   Print desktop session internal info\n"+
			"", filepath.Base(os.Args[0]))
}

func main() {
	os.Exit(Run())
}

func Run() int {
	if len(os.Args) <= 1 || os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help" {
		Usage()
		return 0
	}

	cmd := os.Args[1]
	// Commands that don't require a desktop session
	switch cmd {
	case "profiles":
		return RunProfiles(os.Args[1:])
	case "edit":
		return RunEdit(os.Args[1:])
	}

	var err error
	session, err = GetDesktopSession()
	if err != nil {
		fmt.Println("Error opening desktop session:", err)
		return 1
	}
	defer session.Close()

	// Commands that require a desktop session
	switch cmd {
	case "state":
		return RunState(os.Args[1:])
	case "resources":
		return RunResources(os.Args[1:])
	case "save":
		return RunSave(os.Args[1:])
	case "debuginfo":
		return RunDebugInfo(os.Args[1:])
	}

	fmt.Printf("Invalid command '%s'\n", cmd)
	return 1
}

func GetDesktopSession() (common.DesktopSession, error) {
	// Detect current desktop session
	session := os.Getenv("DESKTOP_SESSION")
	switch session {
	case "gnome":
		return mutter.GetDesktopSession()
	}
	return nil, fmt.Errorf("unsupported desktop session '%s'", session)
}

func RunDebugInfo(args []string) int {
	session.DebugInfo(os.Stdout)
	return 0
}

func RunResources(args []string) int {
	res, err := session.Resources()
	if err != nil {
		fmt.Println("Error getting monitor resources:", err)
		return 1
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(&res)
	return 0
}

func RunState(args []string) int {
	st, err := session.ScreenStates()
	if err != nil {
		fmt.Println("Error getting current monitor layout:", err)
		return 1
	}
	state := common.State{
		Monitors: st,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(&state)
	return 0
}

func RunProfiles(args []string) int {
	files, err := os.ReadDir(filepath.Join(common.GetConfigDir(), "profiles"))
	if err != nil {
		fmt.Println("Error reading profiles:", err)
		return 1
	}

	profiles := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" || len(f.Name()) <= 5 {
			continue
		}
		profiles = append(profiles, f.Name()[:len(f.Name())-5])
	}

	if len(profiles) == 0 {
		fmt.Println("No saved profiles")
		return 0
	}

	for _, p := range profiles {
		fmt.Println(p)
	}
	return 0
}

func RunSave(args []string) int {
	if len(args) != 2 {
		fmt.Println("Give the profile a name")
		return 1
	}

	name := args[1]

	monitors, err := session.ScreenStates()
	if err != nil {
		fmt.Println("Error getting current layout:", err)
		return 1
	}

	profile := common.Profile{
		Monitors: monitors,
	}

	file, err := os.Create(filepath.Join(common.GetConfigDir(), "profiles",
		fmt.Sprintf("%s.json", name)))
	if err != nil {
		fmt.Println("Error creating profile:", err)
		return 1
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	err = enc.Encode(profile)
	if err != nil {
		fmt.Println("Error saving profile:", err)
		return 1
	}

	return 0
}

func RunEdit(args []string) int {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Println("$EDITOR not set")
		return 1
	}

	if len(args) != 2 {
		fmt.Println("Specify which profile to edit")
		return 1
	}

	fp := filepath.Join(common.GetConfigDir(), "profiles",
		fmt.Sprintf("%s.json", args[1]))

	cmd := exec.Command(editor, fp)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running editor:", err)
		return 1
	}

	return 0
}
