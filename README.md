# unifi - Local UniFi Controller CLI

[![CI](https://github.com/dl-alexandre/Local-UniFi-CLI/actions/workflows/ci.yml/badge.svg)](https://github.com/dl-alexandre/Local-UniFi-CLI/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/dl-alexandre/Local-UniFi-CLI/branch/main/graph/badge.svg)](https://codecov.io/gh/dl-alexandre/Local-UniFi-CLI)
[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/Local-UniFi-CLI)](https://goreportcard.com/report/github.com/dl-alexandre/Local-UniFi-CLI)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A command-line interface for local UniFi Network controllers.

## Overview

`unifi` provides local management of UniFi Network installations via the local controller API. This tool connects directly to your UniFi Controller (Dream Machine, Cloud Key, or self-hosted) and manages:

- **Sites** - List and manage network sites
- **Devices** - View access points, switches, and gateways
- **Clients** - Monitor connected devices

For cloud-based management across multiple sites, use the separate `usm` CLI.

## Installation

### Homebrew
```bash
brew tap dl-alexandre/tap
brew install unifi
```

### Manual
Download from GitHub Releases.

## Quick Start

1. **Configure**:
```bash
unifi init
# Or set environment variables:
export UNIFI_BASE_URL="https://192.168.1.1"
export UNIFI_USERNAME="admin"
export UNIFI_PASSWORD="your-password"
```

2. **List sites**:
```bash
unifi sites list
```

3. **List devices**:
```bash
unifi devices list --site default
```

4. **List clients**:
```bash
unifi clients list --site default
```

## Commands

### `unifi init`
Interactive configuration setup.

### `unifi ping`
Test connectivity to the controller.

```bash
unifi ping  # Verifies connection and authentication
```

**Output:**
```
✓ Successfully connected to UniFi controller at https://192.168.1.1
✓ Authentication successful
✓ Found 1 site(s)

Sites available:
  - default (site1)
```

### `unifi sites list`
List all sites on the controller.

### `unifi sites stats [<site>]`
Show site health and statistics.

```bash
unifi sites stats              # Show stats for first available site
unifi sites stats default      # Show stats for specific site
```

**Output:**
```
Site: Default (site1)
Description: Main office

Devices:
  Access Points: 3
  Switches:      2
  Gateways:      1
  Total:         6

Connected Clients: 42

Health Status:
Subsystem    Status    Adopted    Disconnected    Pending
wlan         ✓ ok      3          0               0
lan          ✓ ok      2          0               0
guest        ✓ ok      0          0               0
```

### `unifi devices list`
List all devices (APs, switches, gateways).

```bash
unifi devices list                    # Use first available site
unifi devices list --site <site-id>   # Specific site
```

### `unifi devices adopt <mac>`
Adopt a pending device by MAC address.

```bash
# Adopt a new UniFi device that's in "pending" state
unifi devices adopt aa:bb:cc:dd:ee:ff

# Adopt on a specific site
unifi devices adopt aa:bb:cc:dd:ee:ff --site default
```

### `unifi devices provision <device-id>`
Provision (push configuration to) a device. Use this after making configuration changes in the UniFi controller UI to force the device to update.

```bash
# Provision a device (trigger config push)
unifi devices provision 5f8d9a2b3c4d5e6f7a8b9c0d

# Provision on a specific site
unifi devices provision 5f8d9a2b3c4d5e6f7a8b9c0d --site default
```

### `unifi devices restart <mac>`
Restart a device by MAC address. Useful for troubleshooting or applying updates.

```bash
# Restart an access point
unifi devices restart aa:bb:cc:dd:ee:ff

# Restart on a specific site
unifi devices restart aa:bb:cc:dd:ee:ff --site default
```

**Note:** The device will reboot and may be unavailable for a few minutes.

### `unifi networks list`
List all networks/VLANs configured on the controller.

```bash
unifi networks list                    # Use first available site
unifi networks list --site <site-id>   # Specific site
```

### `unifi networks create <name>`
Create a new network/VLAN.

```bash
# Create a basic corporate VLAN
unifi networks create "VLAN30" --vlan 30 --subnet 192.168.30.0/24

# Create a guest network
unifi networks create "Guest WiFi" --vlan 50 --subnet 192.168.50.0/24 --purpose guest

# Create a VLAN-only network (no gateway)
unifi networks create "IoT VLAN" --vlan 100 --purpose vlan-only
```

**Options:**
- `--vlan`: VLAN ID (1-4094, default: 1)
- `--subnet`: IP subnet in CIDR notation (e.g., `192.168.10.0/24`)
- `--purpose`: Network purpose - `corporate`, `guest`, or `vlan-only` (default: corporate)
- `--dhcp`: Enable DHCP server (default: true)
- `--guest`: Mark as guest network

### `unifi clients list`
List connected clients.

```bash
unifi clients list                    # Use first available site
unifi clients list --site <site-id>   # Specific site
```

### `unifi clients block <mac>`
Block a client by MAC address. The client will be disconnected and prevented from reconnecting.

```bash
# Block a problematic client
unifi clients block aa:bb:cc:dd:ee:f1

# Block on a specific site
unifi clients block aa:bb:cc:dd:ee:f1 --site default
```

### `unifi clients unblock <mac>`
Unblock a previously blocked client. The client will be able to reconnect to the network.

```bash
# Unblock a client
unifi clients unblock aa:bb:cc:dd:ee:f1

# Unblock on a specific site
unifi clients unblock aa:bb:cc:dd:ee:f1 --site default
```

### `unifi settings list`
List controller and site settings.

```bash
# List all settings for the default site
unifi settings list

# List settings for a specific site
unifi settings list --site default

# List only network-related settings
unifi settings list --category network
```

**Example output:**
```
Settings for site: default

site_name                      string        Default
auto_upgrade                   bool          true
wifi_enabled                   bool          true
guest_portal                   bool          false
radius_enabled                 bool          false
syslog_enabled                 bool          true
...
```

### `unifi settings get <key>`
Get the value of a specific setting.

```bash
# Get the site name
unifi settings get site_name

# Get auto-upgrade setting
unifi settings get auto_upgrade

# Get a setting from a specific site
unifi settings get wifi_enabled --site default
```

**Example output:**
```
Setting: wifi_enabled
Type:    bool
Value:   true

Category: wireless
```

### `unifi users list`
List local UniFi users and administrators.

```bash
unifi users list
```

**Example output:**
```
Local Users (2):

admin                Admin User                admin (admin)   enabled
readonly             Read Only                 readonly        enabled
  Email: readonly@example.com
guestuser            Guest Access              readonly        enabled
```

### `unifi users create <name>`
Create a new local UniFi user.

```bash
# Create a readonly user
unifi users create "John Doe" --username johndoe --email john@example.com --password securepass123

# Create an admin user
unifi users create "Admin User" --username admin2 --email admin2@example.com --password adminpass123 --admin

# Create a user with specific role
unifi users create "Viewer" --username viewer --password viewpass123 --role readonly
```

**Options:**
- `--username`: Username for login (required)
- `--email`: Email address (optional)
- `--password`: Password for the user (required)
- `--role`: User role - `admin` or `readonly` (default: readonly)
- `--admin`: Grant admin privileges (sets role to admin)

### `unifi users delete <username-or-id>`
Delete a local UniFi user. You can specify either the username or the user ID.

```bash
# Delete by username (with confirmation prompt)
unifi users delete johndoe

# Delete by user ID (with confirmation prompt)
unifi users delete 507f1f77bcf86cd799439011

# Delete without confirmation (force)
unifi users delete johndoe --force
```

**Options:**
- `--force, -f`: Skip confirmation prompt

### `unifi users set-password <username>`
Set a new password for an existing user.

```bash
# Set password by username
unifi users set-password johndoe --new-password newsecurepass123

# Set password by user ID
unifi users set-password 507f1f77bcf86cd799439011 --new-password newsecurepass123
```

**Options:**
- `--new-password`: New password for the user (required)

### `unifi backups list`
List available controller backups.

```bash
unifi backups list
```

**Example output:**
```
Backups (3):

Filename                       Size         Date                 Type       Encrypted
------------------------------------------------------------------------------------------
backup_2024-01-15.unf          10.00 MB     2024-01-15 08:30     backup     no
autobackup_2024-01-14.unf      10.00 MB     2024-01-14 02:00     autobackup yes
backup_2024-01-13.unf          9.50 MB      2024-01-13 14:22     backup     no
```

### `unifi backups create`
Trigger a manual backup of the controller.

```bash
# Create unencrypted backup
unifi backups create

# Create encrypted backup
unifi backups create --encrypt
```

**Options:**
- `--encrypt`: Encrypt the backup file

### `unifi backups download <filename-or-id>`
Download a backup file to your local machine.

```bash
# Download by filename
unifi backups download backup_2024-01-15.unf

# Download by backup ID
unifi backups download 507f1f77bcf86cd799439011

# Download to specific path
unifi backups download backup_2024-01-15.unf --output /path/to/backup.unf
```

**Options:**
- `--output, -o`: Output file path (default: backup filename in current directory)

### `unifi backups restore <filename-or-id>`
Restore the controller from a backup file. **WARNING**: This will replace all current settings.

```bash
# Restore by filename (with confirmation)
unifi backups restore backup_2024-01-15.unf

# Restore by backup ID (with confirmation)
unifi backups restore 507f1f77bcf86cd799439011

# Restore without confirmation (force)
unifi backups restore backup_2024-01-15.unf --force
```

**Options:**
- `--force, -f`: Skip confirmation prompt

### `unifi firmware list`
List firmware status for all devices showing current and available versions.

```bash
unifi firmware list
```

**Example output:**
```
Firmware Status (5 devices):

Device               Model              Current         Latest          Status
-------------------------------------------------------------------------------------
AP-LivingRoom        UAP-AC-Pro         6.6.55          6.6.55          up-to-date
Switch-Basement      USW-Pro-24         6.6.53          6.6.55          upgrade-available
AP-Bedroom           UAP-nanoHD         6.6.55          6.6.55          up-to-date
Router-Office        UDM-Pro            3.2.15          3.2.17          upgrade-available
AP-Garage            UAP-AC-Lite        6.6.50          6.6.55          upgrade-available
```

### `unifi firmware upgrade <mac-or-id>`
Upgrade firmware for a specific device. You can specify either the MAC address or device ID.

```bash
# Upgrade to latest firmware (using MAC address)
unifi firmware upgrade aa:bb:cc:dd:ee:ff

# Upgrade using device ID
unifi firmware upgrade 507f1f77bcf86cd799439011

# Upgrade to specific version
unifi firmware upgrade aa:bb:cc:dd:ee:ff --version 6.6.60
```

**Options:**
- `--version`: Specific firmware version to upgrade to (default: latest available)

### `unifi port list`
List all switch ports with their current status, PoE state, assigned profile, and VLAN.

```bash
unifi port list
```

**Example output:**
```
Switch Ports:

Device: Switch-1 (aa:bb:cc:dd:ee:01)
Port   Name                 Status     Speed    Profile      PoE      VLAN
--------------------------------------------------------------------------------
1      Office AP            up         1G       All          on       1
2      Server               up         10G      Servers      off      10
3      Printer              down       -        Default      off      1
4      IP Phone             up         100M     VoIP         auto     20
```

### `unifi port set <port-id> --profile=...`
Configure switch port settings. Port ID format is `deviceID/portIndex`.

```bash
# Assign port profile
unifi port set device1/1 --profile=profile2

# Enable PoE on a port
unifi port set device1/4 --poe=auto

# Disable PoE
unifi port set device1/4 --poe=off

# Enable the port
unifi port set device1/3 --enable

# Disable the port
unifi port set device1/3 --disable
```

**Options:**
- `--profile`: Port profile ID to assign
- `--poe`: PoE mode (`auto`, `passthrough`, `off`)
- `--enable`: Enable the port
- `--disable`: Disable the port

### `unifi hotspot list`
List all hotspot guests with their connection status and authorization details.

```bash
unifi hotspot list
```

**Example output:**
```
Hotspot Guests (2):

MAC                  IP              Name/Email           Status      AP              Duration
---------------------------------------------------------------------------------------------------------
aa:bb:cc:dd:ee:01   192.168.10.100  John Doe (john@e...  authorized  AP-Lobby        1440m
aa:bb:cc:dd:ee:02   192.168.10.101  -                    pending     AP-Lobby        -
```

### `unifi hotspot authorize <mac>`
Authorize a guest device for hotspot access.

```bash
# Authorize for default 24 hours (1440 minutes)
unifi hotspot authorize aa:bb:cc:dd:ee:01

# Authorize for 1 hour
unifi hotspot authorize aa:bb:cc:dd:ee:01 --duration 60

# Authorize for 7 days
unifi hotspot authorize aa:bb:cc:dd:ee:01 --duration 10080
```

**Options:**
- `--duration`: Authorization duration in minutes (default: 1440 = 24 hours)

### `unifi hotspot unauthorize <mac>`
Revoke authorization for a guest device, removing their access to the network.

```bash
unifi hotspot unauthorize aa:bb:cc:dd:ee:01
```

### `unifi hotspot kick <mac>`
Immediately disconnect a guest device from the network without revoking their authorization.

```bash
unifi hotspot kick aa:bb:cc:dd:ee:01
```

### `unifi firewall list`
List all firewall rules configured on the controller.

```bash
unifi firewall list                    # Use first available site
unifi firewall list --site <site-id> # Specific site
```

### `unifi firewall create <name>`
Create a new firewall rule.

```bash
# Allow SSH access from WAN
unifi firewall create "Allow SSH" --action accept --protocol tcp --dst-port 22 --ruleset WAN_IN

# Block guest network from accessing LAN
unifi firewall create "Block Guest to LAN" --action drop --src-address 192.168.10.0/24 --dst-address 192.168.1.0/24 --ruleset GUEST_IN

# Allow HTTP/HTTPS from anywhere
unifi firewall create "Allow HTTP" --action accept --protocol tcp --dst-port 80 --ruleset WAN_IN
unifi firewall create "Allow HTTPS" --action accept --protocol tcp --dst-port 443 --ruleset WAN_IN
```

**Options:**
- `--action`: Rule action - `accept`, `drop`, or `reject` (default: accept)
- `--protocol`: Protocol - `all`, `tcp`, `udp`, or `icmp` (default: all)
- `--src-address`: Source address in CIDR notation (e.g., `192.168.1.0/24`)
- `--dst-address`: Destination address in CIDR notation (e.g., `0.0.0.0/0`)
- `--dst-port`: Destination port (e.g., `80`, `443`, `22`)
- `--ruleset`: Rule set - `WAN_IN`, `WAN_OUT`, `LAN_IN`, `LAN_OUT`, `GUEST_IN` (default: LAN_IN)
- `--logging`: Enable logging for this rule
- `--description`: Rule description

### `unifi firewall enable <rule-id>`
Enable an existing firewall rule.

```bash
# Enable a disabled rule
unifi firewall enable rule123

# Enable on specific site
unifi firewall enable rule123 --site default
```

### `unifi firewall disable <rule-id>`
Disable an existing firewall rule (without deleting it).

```bash
# Disable a rule
unifi firewall disable rule123

# Disable on specific site
unifi firewall disable rule123 --site default
```

### `unifi firewall delete <rule-id>`
Delete a firewall rule permanently.

```bash
# Delete a rule (with confirmation prompt)
unifi firewall delete rule123

# Force delete without confirmation
unifi firewall delete rule123 --force
```

### `unifi version [--check]`
Show version information.

### `unifi completion <bash|zsh|fish>`
Generate shell completion scripts.

```bash
# Bash
source <(unifi completion bash)
# Or save permanently:
unifi completion bash > /etc/bash_completion.d/unifi

# Zsh
source <(unifi completion zsh)
# Or save to zsh function path:
unifi completion zsh > "${fpath[1]}/_unifi"

# Fish
unifi completion fish > ~/.config/fish/completions/unifi.fish
```

## Example Output

### Sites
```
$ unifi sites list
ID          Name        Description      Devices    Clients
site1       Default     Main office      8          42
guest       Guest       Guest network    2          15
```

### Devices
```
$ unifi devices list
MAC                 Name        Model       Type    Status     IP
aa:bb:cc:dd:ee:ff   AP-Office   U7PG2       uap     adopted    192.168.1.10
11:22:33:44:55:66   Switch-1    US8P150     usw     adopted    192.168.1.20
```

### Networks/VLANs
```
$ unifi networks list
ID       Name        Purpose    VLAN   Subnet              Enabled
net1     LAN         corporate  -      192.168.1.0/24      Yes
net2     Guest       guest      10     192.168.10.0/24     Yes
net3     IoT         corporate  20     192.168.20.0/24     Yes
```

### Firewall Rules
```
$ unifi firewall list
ID       Name                  Action    Protocol   Source           Destination   Port   Rule Set
rule1    Allow HTTP            accept    tcp        any              any           80     WAN_IN
rule2    Allow HTTPS           accept    tcp        any              any           443    WAN_IN
rule3    Allow SSH             accept    tcp        any              any           22     WAN_IN
rule4    Block Guest to LAN    drop      all        192.168.10.0/24  192.168.1.0/24  any  GUEST_IN
```

### Clients (table format)
```
$ unifi clients list
MAC                 Name        IP              Type      Connected To
aa:bb:cc:dd:ee:f1   iPhone      192.168.1.100   Wireless  aa:bb:cc:dd:ee:ff
aa:bb:cc:dd:ee:f2   Desktop     192.168.1.101   Wired     -
```

### Clients (JSON format)
```
$ unifi clients list --format json
[
  {
    "mac": "aa:bb:cc:dd:ee:f1",
    "name": "iPhone",
    "ip": "192.168.1.100",
    "is_wired": false,
    "signal": -45,
    "ap_mac": "aa:bb:cc:dd:ee:ff"
  }
]
```

## Configuration

Config file: `~/.config/unifi/config.yaml`

```yaml
api:
  base_url: https://192.168.1.1
  timeout: 30

auth:
  username: admin

output:
  format: table
  color: auto
```

**Note:** Passwords are NOT stored in config. Use environment variables or flags.

## Environment Variables

- `UNIFI_BASE_URL` - Controller URL
- `UNIFI_USERNAME` - Username
- `UNIFI_PASSWORD` - Password
- `UNIFI_FORMAT` - Output format
- `UNIFI_COLOR` - Color mode

## UniFi OS Support

### UniFi OS Limitations

The following features are not available on UniFi OS controllers (UDM, UDM Pro, Cloud Key Gen2+, etc.) due to API differences:

- **backups** - UniFi OS handles backups via the filesystem rather than the Network API. Use SSH or the UniFi OS interface directly for backup operations.
- **firmware list** - UniFi OS uses a different firmware management system. Device upgrades are handled through the standard device management commands.

These limitations only affect UniFi OS controllers. Traditional software controllers (self-hosted, Cloud Key Gen1) support all features.

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Authentication failure
- `4` - Validation error
- `5` - Rate limited
- `6` - Network error

## Shell Completions

Tab completion is available for bash, zsh, and fish shells.

### Installation

**Bash:**
```bash
# Temporary (current session)
source <(unifi completion bash)

# Permanent (system-wide)
sudo unifi completion bash > /etc/bash_completion.d/unifi

# Permanent (user only)
unifi completion bash >> ~/.bashrc
```

**Zsh:**
```bash
# Temporary (current session)
source <(unifi completion zsh)

# Permanent (requires fpath setup)
mkdir -p ~/.local/share/zsh/site-functions
unifi completion zsh > ~/.local/share/zsh/site-functions/_unifi
# Add to ~/.zshrc: fpath+=(~/.local/share/zsh/site-functions)
```

**Fish:**
```bash
# Permanent
unifi completion fish > ~/.config/fish/completions/unifi.fish
```

### Features

- Command completion: `unifi si<TAB>` → `unifi sites`
- Flag completion: `unifi sites list --<TAB>` shows available flags
- Enum completion: `unifi --format <TAB>` shows `table` and `json`
- Subcommand completion: `unifi sites <TAB>` shows `list`

## Related Tools

- `usm` - For cloud-based UniFi Site Manager API

## License

MIT
