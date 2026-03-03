# Skill: unifi-cli

# Local UniFi Controller CLI

This skill provides guidance for the `unifi` CLI tool, a command-line interface for local UniFi Network controllers.

## Overview

`unifi` connects directly to your local UniFi Controller (Dream Machine, Cloud Key, or self-hosted) to manage sites, devices, and clients. For cloud-based management across multiple sites, use the separate `usm` CLI.

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
unifi devices list
```

4. **List clients**:
```bash
unifi clients list
```

## Available Commands

### unifi init
Interactive configuration setup. Creates `~/.config/unifi/config.yaml`.

**Use when:**
- First-time setup
- Changing controller URL or username

### unifi sites list
List all sites on the controller.

**Use when:**
- Discovering available sites
- Finding site IDs for other commands

### unifi devices list
List all UniFi devices (APs, switches, gateways).

**Use when:**
- Viewing network hardware
- Checking device status and adoption state
- Finding device MAC addresses

**Flags:**
- `--site=<id>` - Specific site (default: first available)

### unifi clients list
List connected clients.

**Use when:**
- Viewing connected devices
- Checking wireless signal strength
- Finding client MAC addresses

**Flags:**
- `--site=<id>` - Specific site (default: first available)

### unifi version [--check]
Show version information.

## Configuration

Config file: `~/.config/unifi/config.yaml`

**Note:** Passwords are NOT stored in config. Use environment variables or flags.

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

## Environment Variables

| Variable | Description |
|----------|-------------|
| `UNIFI_BASE_URL` | Controller URL (e.g., https://192.168.1.1) |
| `UNIFI_USERNAME` | Username for authentication |
| `UNIFI_PASSWORD` | Password for authentication |
| `UNIFI_FORMAT` | Output format: table, json |
| `UNIFI_COLOR` | Color mode: auto, always, never |
| `UNIFI_NO_HEADERS` | Disable table headers |
| `UNIFI_CONFIG` | Path to config file |

## Authentication

Local UniFi Controllers use session-based authentication:
1. Client sends username/password to `/api/auth/login`
2. Server sets session cookie
3. Client uses cookie for subsequent requests
4. Automatic re-login when session expires

**Security notes:**
- Use local admin accounts (not cloud/SSO accounts)
- Passwords never stored in config
- Session cookies handled automatically

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication failure |
| 4 | Validation error |
| 5 | Rate limited |
| 6 | Network error |

## Common Workflows

### Check all devices across all sites
```bash
for site in $(unifi sites list --output json | jq -r '.[].name'); do
  echo "=== Site: $site ==="
  unifi devices list --site "$site"
done
```

### Export device list to JSON
```bash
unifi devices list --output json > devices.json
```

### Find client by IP
```bash
unifi clients list --output json | jq '.[] | select(.ip == "192.168.1.100")'
```

### Check adoption status
```bash
unifi devices list --output json | jq '.[] | select(.adopted == false)'
```

## API Reference

Local UniFi Controller API endpoints:
- `GET /api/self/sites` - List sites
- `GET /api/s/{site}/stat/device` - List devices
- `GET /api/s/{site}/stat/sta` - List clients
- `GET /api/s/{site}/stat/health` - Site health

## Related Tools

- `usm` - For cloud-based UniFi Site Manager API (multi-site overview)

## Troubleshooting

**Connection refused:**
- Verify controller URL is correct
- Check if controller is running
- Ensure you're using HTTPS (not HTTP)

**Authentication failed:**
- Use local admin account (not UniFi cloud account)
- Verify username/password
- Check if 2FA is enabled (not supported)

**Certificate errors:**
- Self-signed certificates are common for local controllers
- Use `curl -k` equivalent or add certificate to trust store
