#!/bin/bash
branch=$(git rev-parse --abbrev-ref HEAD)
for remote in $(git remote)
do
	status=$(git rev-list --left-right --count HEAD...$remote/$branch 2>/dev/null)
	(( !$? )) || status="0	0"
	IFS='	' read -ra changes <<< "$status"
	up=${changes[0]}
	down=${changes[1]}
	if [[ $up -eq 0 ]] && [[ $down -eq 0 ]]; then continue; fi
	echo "Remote $remote: ⬆️  ${changes[0]} ⬇️  ${changes[1]}"
done
