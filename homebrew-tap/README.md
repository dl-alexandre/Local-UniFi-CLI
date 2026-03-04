# Dalton Alexandre's Homebrew Tap

This is a Homebrew tap containing formulas for UniFi CLI tools.

## Available Formulas

### `unifi` - UniFi Controller CLI

Command-line interface for managing UniFi network controllers locally.

**Installation:**
```bash
brew tap dl-alexandre/unifi-cli
brew install unifi
```

**Features:**
- Connect to local UniFi controllers
- Manage devices, clients, and networks
- Configuration management
- Backup and restore functionality

## Usage

```bash
# Show help
unifi --help

# Initialize configuration
unifi config init

# Connect to your controller
unifi connect <controller-url>

# List devices
unifi devices list
```

## Updates

To update to the latest version:
```bash
brew upgrade unifi
```

## Uninstallation

```bash
brew uninstall unifi
brew untap dl-alexandre/unifi-cli
```

## Development

This tap is maintained by [Dalton Alexandre](https://github.com/dl-alexandre).

## License

These formulas are provided under the MIT License.
