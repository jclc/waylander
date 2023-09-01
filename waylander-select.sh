#!/bin/env bash

IFS="
"
profiles=$(waylander profiles)

if [[ "${#profiles[@]}" -eq 1 && -z "${profiles[0]}" ]]; then
	zenity --info --title="waylander" --text="No saved profiles."
	exit
fi

selected=$(zenity --list --title="waylander" --text="Select monitor layout" \
	--column="Profile" $profiles)
if [[ $? -ne 0 ]]; then
	echo "Aborted."
	exit 1
fi

msg=$(waylander apply "$selected")
if [[ $? -ne 0 ]]; then
	zenity --notification --text="waylander error\n${msg}" --icon=error
fi
