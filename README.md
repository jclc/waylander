# Waylander

A utility for managing monitor layouts on Wayland. Inspired by `xrandr`, `waylander` allows you to save, edit and apply different layouts with a single command.

Under construction.

# Installation (with Go)
```
go install github.com/jclc/waylander/cmd/waylander@latest
```

# Basic usage
You can save the current screen layout as a profile with `waylander save <profile>`. Profiles can then be applied with `waylander apply <profile>`.

Print the full command set with `waylander help`.

## Motivation
I frequently switch my monitor setup between my desktop monitor and TV. Previously I used some `xrandr` scripts, but switched to Wayland some time ago. I wasn't happy with the existing replacements for Wayland so I decided to make something better for myself. I may keep adding extra features to this if I feel like it.

## Supported platforms (so far)
- Gnome Wayland

## Todo list
- [X] Basic documentation
- [X] Ensure cloning works properly
- [ ] KDE support

## Stretch goals
- [ ] GUI
- [ ] Audio devices
