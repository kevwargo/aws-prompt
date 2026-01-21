{{.MainCmd}}()
{
    case "$1" in
        {{.SourcableCommands}})
            . <({{.PreserveStdoutEnv}}=1 {{.RootCmd}} {{.MainCmd}} "$@")
            ;;
        {{.CompletionCommands}})
            local compcmd
            compcmd="$1"
            shift
            {{.RootCmd}} "$compcmd" {{.MainCmd}} "$@"
            ;;
        *)
            {{.RootCmd}} {{.MainCmd}} "$@"
            ;;
    esac
}

. <({{.RootCmd}} bash-completion)
. <({{.RootCmd}} bash-completion {{.MainCmd}})

if [[ "${PROMPT_COMMAND}" != *"{{.MainCmd}} {{.PS1Cmd}}"* ]]; then
    PROMPT_COMMAND+=' ; {{.MainCmd}} {{.PS1Cmd}}'
fi
