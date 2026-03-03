# unifi - Local UniFi Controller CLI

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

### `unifi sites list`
List all sites on the controller.

### `unifi devices list`
List all devices (APs, switches, gateways).

```bash
unifi devices list                    # Use first available site
unifi devices list --site <site-id>   # Specific site
```

### `unifi clients list`
List connected clients.

```bash
unifi clients list                    # Use first available site
unifi clients list --site <site-id>   # Specific site
```

### `unifi version [--check]`
Show version information.

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

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Authentication failure
- `4` - Validation error
- `5` - Rate limited
- `6` - Network error

## Related Tools

- `usm` - For cloud-based UniFi Site Manager API

## License

MIT
