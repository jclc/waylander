#!/bin/env bash

name=$(zenity --forms --title="waylander" --text="Save the monitor layout" --add-entry="Profile name")
if [[ $? -ne 0 ]]; then
	echo "Aborted."
	exit 1
fi

msg=$(waylander save "$name")
if [[ $? -ne 0 ]]; then
	zenity --notification --text="waylander error\n${msg}" --icon=error
fi
