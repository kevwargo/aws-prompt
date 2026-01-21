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

_awsp_ps1()
{
    . <({{.PreserveStdoutEnv}}=1 {{.RootCmd}} {{.PS1Cmd}})
}

. <({{.RootCmd}} bash-completion)
. <({{.RootCmd}} bash-completion {{.MainCmd}})

if [[ "${PROMPT_COMMAND}" != *"_awsp_ps1"* ]]; then
    PROMPT_COMMAND+=' ; _awsp_ps1'
fi
