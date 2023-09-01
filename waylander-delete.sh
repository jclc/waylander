#!/bin/env bash

IFS="
"
profiles=($(waylander profiles))

if [[ "${#profiles[@]}" -eq 1 && -z "${profiles[0]}" ]]; then
	zenity --info --title="waylander" --text="No saved profiles."
	exit
fi

profiles_zen=()
for p in "${profiles[@]}"; do
	profiles_zen+=("FALSE")
	profiles_zen+=("$p")
done

delete=($(zenity --list --checklist --title="waylander" --separator="\n" \
	--text="Select profiles to delete" \
	--column="Delete" --column="Profile" "${profiles_zen[@]}"))
if [[ $? -ne 0 ]]; then
	echo "Aborted."
	exit 1
fi

for profile in "${delete[@]}"; do
	msg=$(waylander delete "$profile")
	if [[ $? -ne 0 ]]; then
		zenity --notification --text="waylander error\n${msg}" --icon=error
	fi
done
