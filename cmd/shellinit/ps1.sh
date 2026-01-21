if [[ -n "${ps1_part_2}" && "${PS1}" == *"${ps1_part_2}" ]]; then
    awsp_ps1_part="{{ . }}"
    awsp_ps1_part="${awsp_ps1_part:+$awsp_ps1_part }"
    PS1="${PS1%"${ps1_part_2}"}${awsp_ps1_part}${ps1_part_2}"
    unset awsp_ps1_part
fi
