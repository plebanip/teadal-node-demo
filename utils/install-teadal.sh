#!/usr/bin/env bash
set -euo pipefail

# Function to display usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Mandatory options:"
    echo "  -u <repo_url>     Specify the repoURL"
    echo "Options:"
    echo "  -d <repo_dir>     Specify the directory where to clone the repo (default:teadal.node)"
    echo "  -b <branch>       Specify a branch (default: HEAD)"
    echo "  -h                Display this help message"
}

# Global variables
repo_dir="./teadal.node" # default Directory with the repo clone
repo_url="" # Url of the repo must be specified
repo_branch="HEAD"       # Branch of the repo

while getopts "d:u:b:h" opt; do
   case $opt in
      d) repo_dir="$OPTARG" ;;
      u) repo_url="$OPTARG" ;;
      b) repo_branch="$OPTARG" ;;
      h)
          usage
          exit 0
          ;;
      ?)
          error "Invalid option: $OPTARG"
          exit 1
          ;;
    esac
done


# Check mandatory arguments
if [[ -z "$repo_url" ]]; then
    echo "Error: -u <repo> are required"
    exit 1
fi

# Assuming this script is running from project root, otherwise adjust accordingly using these environment variables
FLAKE_DIR="${FLAKE_DIR:-$(pwd)/nix}"
UTILS_DIR="${UTILS_DIR:-$(pwd)/utils}"

# Colors and logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
log() { echo -e "${GREEN}[INFO]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1"; }
error() {
    echo -e "${RED}[ERROR]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" >&2
    exit 1
}

log "Running script with ${repo_url}, ${repo_dir}, ${repo_branch}"

#clone the repo
if [ ! -d "$repo_dir" ]; then

  if [[ "$repo_branch" == "HEAD" ]]; then
    git clone "$repo_url" "$repo_dir" ||  error "Problems in cloning the repo."
  else
    git clone -b "$repo_branch" "$repo_url" "$repo_dir" ||  error "Problems in cloning the repo."
  fi
else
  log "Directory ${repo_dir} already exists" 
  read -p "Do you want to lauch Teadal from the existing repo? (y/n): " answer

    case "$answer" in
        y|Y ) echo "Continuing..."; cd "${repo_dir}"/nix; nix run .#teadal-deployment || error "Failed to install Teadal node"; exit 1 ;;
        n|N ) echo "Aborting."; exit 1 ;;
        * ) echo "Invalid answer. Aborting."; exit 1 ;;
    esac


fi

log "Running script with ${repo_url}, ${repo_dir}, ${repo_branch}"
cd "$repo_dir"

#update the references for argocd
sudo sed -i "s|^    repoURL: .*|    repoURL: ${repo_url}|" "${PWD}"/deployment/mesh-infra/argocd/projects/base/app.yaml
sudo sed -i "s|^    targetRevision: .*|    targetRevision: ${repo_branch}|" "${PWD}"/deployment/mesh-infra/argocd/projects/base/app.yaml


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
    # update the env variables required to run nix
    . ~/.nix-profile/etc/profile.d/nix.sh

    #check the installation
    nix --version || error "Nix installation failed."

    #ask to exit from the shell and re-enter
    log "To make nix properly running you need to logout from the terminal and re-connent"
    log "Then re-run the TEADAL install procedure with"
    log "install-teadal.sh -u <repo_url>"
    log "say 'yes' to use the same directory"
else
    log "Nix is already installed."
    # Be sure that flakes is supportedd
    mkdir -p ~/.config/nix
    echo "experimental-features = nix-command flakes" >~/.config/nix/nix.conf

    log "Starting teadal node generation in nix environment..."
    log "Ready to install TEADAL node"
    log "To install the actual TEADAL node execute the followinf commands: "
    log "  cd nix"
    log "  nix shell"
    log "  nix run .#teadal-deployment"
    #nix run .#teadal-deployment || error "Failed to install Teadal node"

    #log "TEADAL node installed"
    log "Once installted rememeber that to use Teadal you have to operate via nix"
    log "e.g.: firstly open a nix shell 'nix shell'"
    log "      then  look at the current configuration 'kubectl get pod -A'"
fi
