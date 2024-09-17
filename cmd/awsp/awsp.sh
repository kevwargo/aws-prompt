{{.MainCmd}}()
{
    {{.RootCmd}} {{.MainCmd}} "$@" | while true
    do
        IFS='' read -r line
        read_rc=$?

        if [ "$line" = "{{.SourceStart}}" ]; then
            . /dev/stdin
            return
        elif [ $read_rc = 0 ]; then
            echo "$line"
        else
            echo -n "$line"
            return
        fi
    done
}

. <({{.RootCmd}} completion bash)

if [[ "${PROMPT_COMMAND}" != *"{{.MainCmd}} {{.PS1Cmd}}"* ]]; then
    PROMPT_COMMAND+=' ; {{.MainCmd}} {{.PS1Cmd}}'
fi
