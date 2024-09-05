{{ .MainCmd }}()
{
    . <({{ .RootCmd }} {{ .MainCmd }} "$@")
}

if [[ "${PROMPT_COMMAND}" != *"{{ .MainCmd }} {{ .PS1Cmd }}"* ]]; then
    PROMPT_COMMAND+=' ; {{ .MainCmd }} {{ .PS1Cmd }}'
fi
