package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/jclc/waylander/common"
	"github.com/jclc/waylander/mutter"
)

var (
	session common.DesktopSession
)

func Usage() {
	fmt.Println("Waylander -- a Wayland screen management tool.")
	fmt.Println()
	fmt.Printf(
		"Usage: %s <command> [args...]\n"+
			"\n"+
			"Commands:\n"+
			"    help                     Print this help message\n"+
			"    resources                Show currently connected outputs\n"+
			"    state                    Show the current configuration\n"+
			"    profiles                 List saved profiles\n"+
			"    save <profile>           Save current state as a profile\n"+
			"    show <profile>           Show profile\n"+
			"    apply [opts] <profile>   Apply profile\n"+
			"      -persist               Make profile persistent\n"+
			"      -verify                Ask for confirmation\n"+
			"    edit <profile>           Edit profile\n"+
			"    debuginfo                Print desktop session internal info\n"+
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
		return RunProfiles(os.Args[2:])
	case "show":
		return RunShow(os.Args[2:])
	case "edit":
		return RunEdit(os.Args[2:])
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
		return RunState(os.Args[2:])
	case "resources":
		return RunResources(os.Args[2:])
	case "apply":
		return RunApply(os.Args[2:])
	case "save":
		return RunSave(os.Args[2:])
	case "debuginfo":
		return RunDebugInfo(os.Args[2:])
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

func getProfilePath(name string) string {
	return filepath.Join(common.GetConfigDir(), "profiles",
		fmt.Sprintf("%s.json", name))
}

func getProfiles() []string {
	files, err := os.ReadDir(filepath.Join(common.GetConfigDir(), "profiles"))
	if err != nil {
		fmt.Println("Error reading profiles:", err)
		os.Exit(1)
	}

	profiles := make([]string, 0, len(files))
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" || len(f.Name()) <= 5 {
			continue
		}
		profiles = append(profiles, f.Name()[:len(f.Name())-5])
	}

	return profiles
}

func RunDebugInfo(args []string) int {
	err := session.DebugInfo(os.Stdout)
	if err != nil {
		fmt.Printf("Error getting debug info: %s", err)
		return 1
	}
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
	profiles := getProfiles()

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
	if len(args) != 1 || args[0] == "" {
		fmt.Println("Give the profile a name")
		return 1
	}

	monitors, err := session.ScreenStates()
	if err != nil {
		fmt.Println("Error getting current layout:", err)
		return 1
	}

	profile := common.Profile{
		Monitors: monitors,
	}

	file, err := os.Create(getProfilePath(args[0]))
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

func RunShow(args []string) int {
	if len(args) == 0 {
		fmt.Println("Specify a profile to show")
		return 1
	}

	if !slices.Contains(getProfiles(), args[0]) {
		fmt.Printf("Profile '%s' does not exist\n", args[0])
		return 1
	}

	profileData, err := os.ReadFile(getProfilePath(args[0]))
	if err != nil {
		fmt.Println("Error reading profile:", err)
		return 1
	}

	fmt.Print(string(profileData))

	return 0
}

func RunEdit(args []string) int {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		fmt.Println("$EDITOR not set")
		return 1
	}

	if len(args) != 1 {
		fmt.Println("Specify which profile to edit")
		return 1
	}

	cmd := exec.Command(editor, getProfilePath(args[0]))
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

func RunApply(args []string) int {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	var verify, persist bool
	set.BoolVar(&persist, "persist", false,
		"Make profile persistent")
	set.BoolVar(&verify, "verify", false,
		"Ask for confirmation")

	err := set.Parse(args)
	if err != nil {
		return 1
	}
	args = set.Args()

	if len(args) != 1 {
		fmt.Println("Specify which profile to apply")
		return 1
	}

	if !slices.Contains(getProfiles(), args[0]) {
		fmt.Printf("Profile '%s' does not exist\n", args[0])
		return 1
	}

	profileData, err := os.ReadFile(getProfilePath(args[0]))
	if err != nil {
		fmt.Println("Error reading profile:", err)
		return 1
	}

	var profile common.Profile
	err = json.Unmarshal(profileData, &profile)
	if err != nil {
		fmt.Println("Error parsing profile:", err)
		return 1
	}

	err = session.Apply(profile, verify, persist)
	if err != nil {
		fmt.Println("Error applying profile:", err)
		return 1
	}

	return 0
}
