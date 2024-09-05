if [[ -n "${ps1_part_2}" && "${PS1}" == *"${ps1_part_2}" ]]; then
    PS1="${PS1%"${ps1_part_2}"}{{ . }} ${ps1_part_2}"
fi
