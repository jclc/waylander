# Waylander

A utility for managing monitor layouts on Wayland. Inspired by `xrandr`, `waylander` allows you to save, edit and apply different layouts with a single command.

# Installation

## Using Go

Requires Go 1.21.0 or later
```
go install github.com/jclc/waylander/cmd/waylander@latest
```

# Commands

`waylander state`

Show the current screen layout.

---

`waylander resources`

Show all connected outputs.

---

`waylander profiles`

List all saved profiles.

Profiles are stored in `${XDG_CONFIG_HOME-~/.config}/waylander/profiles`.

---

`waylander save <profile>`

Save the current layout into a profile.

---

`waylander apply [-persist] <profile>`

Apply the given profile.

`-persist` will save the layout on the desktop environment.

---

`waylander delete <profile>`

Delete the profile.

---

`waylander edit <profile>`

Open the profile in an editor.

The `$EDITOR` environment variable is used to determine the editor.

---

`waylander help`

Print all available commands.

## Motivation
I frequently switch my monitor setup between my desktop monitor and TV. Previously I used some `xrandr` scripts, but switched to Wayland some time ago. I wasn't happy with the existing replacements for Wayland so I decided to make something better for myself. I may keep adding extra features to this if I feel like it.

## Supported platforms (so far)
- Gnome Wayland
- Gnome Xorg

## Todo list
- [X] Basic documentation
- [X] Ensure cloning works properly
- [ ] KDE support
  - Need to find the proper API first
- [ ] Cross-DE compatible profiles
  - Connector names can differ between desktops

## Stretch goals
- [ ] GUI
- [ ] Optionally include audio devices in profiles
