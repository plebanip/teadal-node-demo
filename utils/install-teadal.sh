#!/usr/bin/env bash
set -euo pipefail

# Assuming this script is running from project root, otherwise adjust accordingly using these environment variables
FLAKE_DIR="${FLAKE_DIR:-$(pwd)/nix}"
UTILS_DIR="${UTILS_DIR:-$(pwd)/utils}"
script="${UTILS_DIR}/teadal-node-generator.sh"

# Colors and logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
log() { echo -e "${GREEN}[INFO]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1"; }
error() {
    echo -e "${RED}[ERROR]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" >&2
    exit 1
}

# Install nix if not present
if ! command -v nix &>/dev/null; then
    log "Nix not found. Installing Nix..."
    curl -L https://nixos.org/nix/install | sh || error "Failed to install Nix."

    # Enable flakes support
    mkdir -p ~/.config/nix
    echo "experimental-features = nix-command flakes" >~/.config/nix/nix.conf

    if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
        . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
    fi

    nix --version || error "Nix installation failed."
else
    log "Nix is already installed."
fi

log "Starting teadal node generation in nix environment..."
nix shell ./nix --command "$script" "$@" || error "Teadal node generation with nix failed."
