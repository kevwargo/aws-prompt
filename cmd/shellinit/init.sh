{{.MainCmd}}()
{
    case "$1" in
        {{.SourcableCommands}})
            local -a payload
            local -a source_lines
            local line

            mapfile payload < <({{.RootCmd}} {{.MainCmd}} "$@")

            for line in "${payload[@]}"; do
                if [ ${#source_lines[@]} -gt 0 -o "$line" = $'{{.SourceStart}}\n' ]; then
                    source_lines+=("$line")
                else
                    echo -n "$line"
                fi
            done

            . <(for line in "${source_lines[@]}"; do echo -n "$line"; done)
            ;;
        {{.CompletionCommands}})
            local compcmd
            compcmd="$1"
            shift
            {{.RootCmd}} "$compcmd" {{.MainCmd}} "$@"
            ;;
    esac
}

. <({{.RootCmd}} bash-completion)
. <({{.RootCmd}} bash-completion {{.MainCmd}})

if [[ "${PROMPT_COMMAND}" != *"{{.MainCmd}} {{.PS1Cmd}}"* ]]; then
    PROMPT_COMMAND+=' ; {{.MainCmd}} {{.PS1Cmd}}'
fi
