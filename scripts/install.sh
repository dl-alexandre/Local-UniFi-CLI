#!/bin/bash
set -e

REPO="dl-alexandre/Local-UniFi-CLI"
BINARY_NAME="unifi"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case $ARCH in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  armv7l) ARCH="arm" ;;
  *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

# Map OS names
case $OS in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

# Get latest release version
echo -e "${YELLOW}Fetching latest release...${NC}"
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo -e "${RED}Failed to fetch latest version${NC}"
  exit 1
fi

echo -e "${GREEN}Latest version: $VERSION${NC}"

# Construct download URL
URL="https://github.com/$REPO/releases/download/${VERSION}/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"

echo -e "${YELLOW}Downloading $BINARY_NAME $VERSION for $OS/$ARCH...${NC}"

# Download to temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

if ! curl -fsL -o "$TMP_DIR/${BINARY_NAME}.tar.gz" "$URL"; then
  echo -e "${RED}Failed to download $URL${NC}"
  exit 1
fi

# Extract
echo -e "${YELLOW}Extracting...${NC}"
tar -xzf "$TMP_DIR/${BINARY_NAME}.tar.gz" -C "$TMP_DIR"

# Install
echo -e "${YELLOW}Installing to /usr/local/bin...${NC}"
if [ -w /usr/local/bin ]; then
  mv "$TMP_DIR/${BINARY_NAME}" /usr/local/bin/
  chmod +x /usr/local/bin/${BINARY_NAME}
else
  sudo mv "$TMP_DIR/${BINARY_NAME}" /usr/local/bin/
  sudo chmod +x /usr/local/bin/${BINARY_NAME}
fi

echo -e "${GREEN}✅ $BINARY_NAME installed successfully!${NC}"
echo ""
echo "Run '$BINARY_NAME --version' to verify"
echo "Run '$BINARY_NAME --help' to see available commands"
