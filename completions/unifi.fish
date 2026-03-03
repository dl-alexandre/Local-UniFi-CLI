# UniFi CLI Fish Completion
# Save to: unifi completion fish > ~/.config/fish/completions/unifi.fish

# Disable file completions for unifi command
complete -c unifi -f

# Global flags
complete -c unifi -l base-url -d "Controller base URL"
complete -c unifi -l username -d "Username for authentication"
complete -c unifi -l password -d "Password for authentication"
complete -c unifi -l timeout -d "Request timeout in seconds"
complete -c unifi -l format -d "Output format" -a "table json"
complete -c unifi -l color -d "Color mode" -a "auto always never"
complete -c unifi -l no-headers -d "Disable table headers"
complete -c unifi -s v -l verbose -d "Enable verbose output"
complete -c unifi -l debug -d "Enable debug output"
complete -c unifi -s c -l config-file -d "Config file path"
complete -c unifi -s h -l help -d "Show help"

# Commands
complete -c unifi -n "__fish_use_subcommand" -a init -d "Interactive configuration setup"
complete -c unifi -n "__fish_use_subcommand" -a ping -d "Test connectivity to controller"
complete -c unifi -n "__fish_use_subcommand" -a sites -d "Manage sites"
complete -c unifi -n "__fish_use_subcommand" -a networks -d "Manage networks/VLANs"
complete -c unifi -n "__fish_use_subcommand" -a devices -d "Manage devices"
complete -c unifi -n "__fish_use_subcommand" -a clients -d "Manage clients"
complete -c unifi -n "__fish_use_subcommand" -a firewall -d "Manage firewall rules"
complete -c unifi -n "__fish_use_subcommand" -a settings -d "Manage controller settings"
complete -c unifi -n "__fish_use_subcommand" -a users -d "Manage local UniFi users"
complete -c unifi -n "__fish_use_subcommand" -a backups -d "Manage controller backups"
complete -c unifi -n "__fish_use_subcommand" -a firmware -d "Manage device firmware"
complete -c unifi -n "__fish_use_subcommand" -a port -d "Manage switch ports"
complete -c unifi -n "__fish_use_subcommand" -a hotspot -d "Manage hotspot guests"
complete -c unifi -n "__fish_use_subcommand" -a version -d "Show version information"
complete -c unifi -n "__fish_use_subcommand" -a completion -d "Generate shell completion scripts"

# Subcommand: sites
complete -c unifi -n "__fish_seen_subcommand_from sites" -a list -d "List all sites"
complete -c unifi -n "__fish_seen_subcommand_from sites" -a stats -d "Show site health and statistics"
complete -c unifi -n "__fish_seen_subcommand_from sites" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from sites" -l help -d "Show help"

# Subcommand: networks
complete -c unifi -n "__fish_seen_subcommand_from networks" -a list -d "List all networks"
complete -c unifi -n "__fish_seen_subcommand_from networks" -a create -d "Create a new network/VLAN"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l name -d "Network name"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l vlan -d "VLAN ID"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l subnet -d "IP subnet (e.g., 192.168.10.0/24)"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l purpose -d "Network purpose" -a "corporate guest vlan-only"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l dhcp -d "Enable DHCP"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l guest -d "Mark as guest network"
complete -c unifi -n "__fish_seen_subcommand_from networks" -l help -d "Show help"

# Subcommand: devices
complete -c unifi -n "__fish_seen_subcommand_from devices" -a list -d "List all devices"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a adopt -d "Adopt a pending device by MAC"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a provision -d "Provision a device"
complete -c unifi -n "__fish_seen_subcommand_from devices" -a restart -d "Restart a device by MAC"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from devices" -l help -d "Show help"

# Subcommand: clients
complete -c unifi -n "__fish_seen_subcommand_from clients" -a list -d "List connected clients"
complete -c unifi -n "__fish_seen_subcommand_from clients" -a block -d "Block a client by MAC address"
complete -c unifi -n "__fish_seen_subcommand_from clients" -a unblock -d "Unblock a client by MAC address"
complete -c unifi -n "__fish_seen_subcommand_from clients" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from clients" -l help -d "Show help"

# Subcommand: firewall
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a list -d "List all firewall rules"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a create -d "Create a new firewall rule"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a enable -d "Enable a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a disable -d "Disable a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -a delete -d "Delete a firewall rule by ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l name -d "Rule name"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l action -d "Rule action" -a "accept drop reject"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l protocol -d "Protocol" -a "all tcp udp icmp"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l src-address -d "Source address (e.g., 192.168.1.0/24)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l dst-address -d "Destination address (e.g., 0.0.0.0/0)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l dst-port -d "Destination port (e.g., 80, 443, 22)"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l ruleset -d "Rule set" -a "WAN_IN WAN_OUT LAN_IN LAN_OUT GUEST_IN"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l logging -d "Enable logging"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l force -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from firewall" -l help -d "Show help"

# Subcommand: settings
complete -c unifi -n "__fish_seen_subcommand_from settings" -a list -d "List controller/site settings"
complete -c unifi -n "__fish_seen_subcommand_from settings" -a get -d "Get a specific setting value"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l site -d "Site ID"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l category -d "Filter by category (network, system, wireless, etc.)"
complete -c unifi -n "__fish_seen_subcommand_from settings" -l help -d "Show help"

# Subcommand: users
complete -c unifi -n "__fish_use_subcommand" -a users -d "Manage local UniFi users"
complete -c unifi -n "__fish_seen_subcommand_from users" -a list -d "List local UniFi users"
complete -c unifi -n "__fish_seen_subcommand_from users" -a create -d "Create a new local user"
complete -c unifi -n "__fish_seen_subcommand_from users" -a delete -d "Delete a local user by ID or username"
complete -c unifi -n "__fish_seen_subcommand_from users" -a set-password -d "Set password for a user"
complete -c unifi -n "__fish_seen_subcommand_from users" -l name -d "Full name"
complete -c unifi -n "__fish_seen_subcommand_from users" -l user -d "Username for login"
complete -c unifi -n "__fish_seen_subcommand_from users" -l email -d "Email address"
complete -c unifi -n "__fish_seen_subcommand_from users" -l user-password -d "Password for the user"
complete -c unifi -n "__fish_seen_subcommand_from users" -l role -d "User role" -a "admin readonly"
complete -c unifi -n "__fish_seen_subcommand_from users" -l is-admin -d "Grant admin privileges"
complete -c unifi -n "__fish_seen_subcommand_from users" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from users" -l new-password -d "New password"
complete -c unifi -n "__fish_seen_subcommand_from users" -l help -d "Show help"
complete -c unifi -n "__fish_seen_subcommand_from users" -l help -d "Show help"

# Subcommand: backups
complete -c unifi -n "__fish_use_subcommand" -a backups -d "Manage controller backups"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a list -d "List available backups"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a create -d "Create a manual backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a download -d "Download a backup file"
complete -c unifi -n "__fish_seen_subcommand_from backups" -a restore -d "Restore from a backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l encrypt -d "Encrypt the backup"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l output -s o -d "Output file path"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l force -s f -d "Skip confirmation prompt"
complete -c unifi -n "__fish_seen_subcommand_from backups" -l help -d "Show help"

# Subcommand: firmware
complete -c unifi -n "__fish_use_subcommand" -a firmware -d "Manage device firmware"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -a list -d "List firmware status for all devices"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -a upgrade -d "Upgrade firmware for a device"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -l version -d "Specific firmware version to upgrade to"
complete -c unifi -n "__fish_seen_subcommand_from firmware" -l help -d "Show help"

# Subcommand: port
complete -c unifi -n "__fish_use_subcommand" -a port -d "Manage switch ports"
complete -c unifi -n "__fish_seen_subcommand_from port" -a list -d "List switch ports with status"
complete -c unifi -n "__fish_seen_subcommand_from port" -a set -d "Configure port settings"
complete -c unifi -n "__fish_seen_subcommand_from port" -l device -d "Filter by device (MAC or ID)"
complete -c unifi -n "__fish_seen_subcommand_from port" -l profile -d "Port profile ID to assign"
complete -c unifi -n "__fish_seen_subcommand_from port" -l poe -d "PoE mode" -a "auto passthrough off"
complete -c unifi -n "__fish_seen_subcommand_from port" -l enable -d "Enable the port"
complete -c unifi -n "__fish_seen_subcommand_from port" -l disable -d "Disable the port"
complete -c unifi -n "__fish_seen_subcommand_from port" -l help -d "Show help"

# Subcommand: hotspot
complete -c unifi -n "__fish_use_subcommand" -a hotspot -d "Manage hotspot guests"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a list -d "List hotspot guests"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a authorize -d "Authorize a guest"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a unauthorize -d "Unauthorize a guest"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -a kick -d "Kick a guest from the network"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -l duration -d "Authorization duration in minutes"
complete -c unifi -n "__fish_seen_subcommand_from hotspot" -l help -d "Show help"

# Subcommand: completion
complete -c unifi -n "__fish_seen_subcommand_from completion" -a bash -d "Generate bash completions"
complete -c unifi -n "__fish_seen_subcommand_from completion" -a zsh -d "Generate zsh completions"
complete -c unifi -n "__fish_seen_subcommand_from completion" -a fish -d "Generate fish completions"
