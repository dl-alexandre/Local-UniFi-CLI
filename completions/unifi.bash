# UniFi CLI Bash Completion
# Source this file: source <(unifi completion bash)
# Or add to ~/.bashrc: source <(unifi completion bash)
# Or save to file: unifi completion bash > /etc/bash_completion.d/unifi

_unifi_completion() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Main commands
    local commands="init ping sites networks devices clients firewall settings users backups firmware port hotspot version completion"
    
    # Global flags
    local global_flags="--base-url --username --password --timeout --format --color --no-headers --verbose --debug --config-file --help"
    
    # Subcommands
    local sites_cmds="list stats"
    local networks_cmds="list create"
    local devices_cmds="list adopt provision restart"
    local clients_cmds="list block unblock"
    local firewall_cmds="list create enable disable delete"
    local settings_cmds="list get"
    local completion_shells="bash zsh fish"
    
    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi
    
    case "${COMP_WORDS[1]}" in
        sites)
            COMPREPLY=( $(compgen -W "list stats --site --help" -- ${cur}) )
            ;;
        networks)
            COMPREPLY=( $(compgen -W "list create --site --name --vlan --subnet --purpose --help" -- ${cur}) )
            ;;
        devices)
            COMPREPLY=( $(compgen -W "list adopt provision restart --site --help" -- ${cur}) )
            ;;
        clients)
            COMPREPLY=( $(compgen -W "list block unblock --site --help" -- ${cur}) )
            ;;
        firewall)
            COMPREPLY=( $(compgen -W "list create enable disable delete --site --name --action --protocol --src-address --dst-address --dst-port --ruleset --help" -- ${cur}) )
            ;;
        settings)
            COMPREPLY=( $(compgen -W "list get --site --category --help" -- ${cur}) )
            ;;
        users)
            COMPREPLY=( $(compgen -W "list create delete set-password --name --user --email --user-password --role --is-admin --force --new-password --help" -- ${cur}) )
            ;;
        backups)
            COMPREPLY=( $(compgen -W "list create download restore --encrypt --output --force --help" -- ${cur}) )
            ;;
        firmware)
            COMPREPLY=( $(compgen -W "list upgrade --version --help" -- ${cur}) )
            ;;
        port)
            COMPREPLY=( $(compgen -W "list set --device --profile --poe --enable --disable --help" -- ${cur}) )
            ;;
        hotspot)
            COMPREPLY=( $(compgen -W "list authorize unauthorize kick --duration --help" -- ${cur}) )
            ;;
        completion)
            COMPREPLY=( $(compgen -W "${completion_shells}" -- ${cur}) )
            ;;
        *)
            COMPREPLY=( $(compgen -W "${global_flags}" -- ${cur}) )
            ;;
    esac
}

complete -F _unifi_completion unifi
