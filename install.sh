#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="arafato/please"
BINARY_NAME="please"
INSTALL_DIR="$HOME/.local/bin"

echo -e "${GREEN}Installing 'please' tool...${NC}"
echo ""

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Normalize architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

# Check if OS is supported
if [ "$OS" != "darwin" ] && [ "$OS" != "linux" ]; then
    echo -e "${RED}Unsupported operating system: $OS${NC}"
    echo "This installer supports macOS (darwin) and Linux only."
    exit 1
fi

echo -e "Detected platform: ${YELLOW}${OS}-${ARCH}${NC}"
echo ""

# Construct the binary name pattern based on your release naming
BINARY_PATTERN="please-${OS}-${ARCH}"

# Get the latest release info from GitHub API
echo "Fetching latest release information..."
RELEASE_INFO=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest")

# Extract the latest version tag
VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo -e "${RED}Failed to fetch latest release version${NC}"
    echo "Please check if releases exist at https://github.com/${REPO}/releases"
    exit 1
fi

echo -e "Latest version: ${GREEN}${VERSION}${NC}"
echo ""

# Find the download URL for the matching binary
DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*${BINARY_PATTERN}-${VERSION}.tar.gz\"" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}Could not find binary for ${OS}-${ARCH}${NC}"
    echo "Looking for: ${BINARY_PATTERN}-${VERSION}.tar.gz"
    echo ""
    echo "Available assets:"
    echo "$RELEASE_INFO" | grep "browser_download_url" | cut -d '"' -f 4
    exit 1
fi

echo -e "Download URL: ${YELLOW}${DOWNLOAD_URL}${NC}"
echo ""

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download the archive
echo "Downloading binary..."
if ! curl -L -o "${TMP_DIR}/please.tar.gz" "$DOWNLOAD_URL"; then
    echo -e "${RED}Failed to download binary${NC}"
    exit 1
fi

# Extract the archive
echo "Extracting archive..."
tar -xzf "${TMP_DIR}/please.tar.gz" -C "$TMP_DIR"

# Find the extracted binary
EXTRACTED_BINARY=$(find "$TMP_DIR" -name "${BINARY_PATTERN}" -type f | head -n 1)

if [ -z "$EXTRACTED_BINARY" ]; then
    echo -e "${RED}Could not find binary in archive${NC}"
    exit 1
fi

# Make it executable
chmod +x "$EXTRACTED_BINARY"

# Ensure installation directory exists
mkdir -p "$INSTALL_DIR"

# Install the binary (no sudo needed for user directory)
echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
mv "$EXTRACTED_BINARY" "${INSTALL_DIR}/${BINARY_NAME}"

# Verify installation
if command -v "$BINARY_NAME" &> /dev/null; then
    echo ""
    echo -e "${GREEN}✓ Successfully installed 'please' ${VERSION}${NC}"
    echo ""
    echo "You can now use the 'please' command:"
    echo -e "  ${YELLOW}$ please --help${NC}"
    echo ""

    # Show version if the binary supports it
    if "$BINARY_NAME" --version &> /dev/null; then
        "$BINARY_NAME" --version
    fi
else
    echo ""
    echo -e "${GREEN}✓ Binary installed to ${INSTALL_DIR}/${BINARY_NAME}${NC}"
    echo ""

    # Check if directory is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}⚠ ${INSTALL_DIR} is not in your PATH${NC}"
        echo ""
        echo "Add the following line to your shell configuration file:"
        echo ""

        # Detect shell and provide appropriate instruction
        if [ -n "$ZSH_VERSION" ]; then
            echo -e "  ${YELLOW}echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc${NC}"
            echo -e "  ${YELLOW}source ~/.zshrc${NC}"
        elif [ -n "$BASH_VERSION" ]; then
            echo -e "  ${YELLOW}echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc${NC}"
            echo -e "  ${YELLOW}source ~/.bashrc${NC}"
        else
            echo -e "  ${YELLOW}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        fi

        echo ""
        echo "Or run this command directly in your current shell:"
        echo -e "  ${YELLOW}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
        echo ""
        echo "Then you can use:"
        echo -e "  ${YELLOW}$ please --help${NC}"
    else
        echo -e "${RED}Installation completed but 'please' command not found${NC}"
        echo "Try opening a new terminal window or running:"
        echo -e "  ${YELLOW}hash -r${NC}"
    fi
fi
