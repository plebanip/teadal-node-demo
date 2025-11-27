#!/bin/bash
set -euo pipefail

# Function to display usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Mandatory options:"
    echo "  -d <repo_dir>     Specify the directory with the repo clone"
    echo "  -u <repo_url>     Specify the repoURL"
    echo "Options:"
    echo "  -b <branch>       Specify a branch"
    echo "  -h                Display this help message"
}

main() {
    parse_options "$@"
    log "Script started with options:\n\trepo_dir=$repo_dir\n\trepo_url=$repo_url\n\tbranch=$branch"
    setup_microk8s "$@" 
    setup_storage
    setup_mesh
    setup_argocd
    log "Script completed successfully."
    exit 0 # TODO: comment before commit until fully tested
}

### Utilities scripts
# Colors and logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

TEADAL_LOG_DIR="${TEADAL_LOG_DIR:-/tmp}"
logfile="$TEADAL_LOG_DIR/install-teadal.log"

log() { echo "${GREEN}[INFO]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a "$logfile"; }
error() { echo "${RED}[ERROR]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a "$logfile" >&2; }
error_exit() {
    error "$1"
    exit "${2:-1}"
}

# Global variables
repo_dir="$(pwd)/.." # Directory with the repo clone
repo_url="$(git config --get remote.origin.url 2>/dev/null || echo '')"
# Url of the repo
branch=""       # Branch of the repo
#hostname_dir="" # Directory with generated storage pv

parse_options() {
    while getopts "d:u:b:h" opt; do
        case $opt in
        d) repo_dir="$OPTARG" ;;
        u) repo_url="$OPTARG" ;;
        b) branch="$OPTARG" ;;
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

    if [ -z "$HOSTNAME" ]; then
        error "HOSTNAME variable not defined"
        exit 1
    fi
}

setup_microk8s() {
    log "Setting up microk8s..."

    # If microk8s is not installed install it
    if ! command -v microk8s &>/dev/null; then
        log "microk8s not found, installing..."
        sudo mkdir -p /var/snap/microk8s/common/ || error_exit "Failed to create /var/snap/microk8s/common/."
        sudo cp "$repo_dir/utils/microk8s-config.yaml" /var/snap/microk8s/common/.microk8s.yaml || error_exit "Failed to copy microk8s configuration file."
        sudo snap install microk8s --classic --channel=1.27/stable || error_exit "Failed to install microk8s."
        sudo usermod -a -G microk8s "$USER"
        #mkdir -p ~/.kube
        #chmod 0700 ~/.kube
        log "User $USER added to microk8s group. You may need to log out and log back in for this to take effect."
        log "Waiting for microk8s to be ready..."
        sudo microk8s status --wait-ready &>/dev/null || error_exit "microk8s is not ready."
        microk8s config | sudo tee ~/.kube/config > /dev/null
        export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
        if ! command -v kubectl &>/dev/null; then
            log "kubectl not found, aliasing it to microk8s.kubectl"
            sudo snap alias microk8s.kubectl kubectl || error_exit "Failed to alias kubectl."
        fi
        sudo microk8s disable ha-cluster --force
        sudo microk8s enable dns
        log "Setup microk8s completed."        
        #exec sg microk8s "$0 $*"
        #newgrp microk8s       
    else 
        #newgrp microk8s 
        log "microk8s found, updating configuration..."
        sudo microk8s start
        sudo snap set microk8s config="$(cat "$repo_dir"/utils/microk8s-config.yaml)"
    fi
}

setup_mesh() {
    log "Setting up mesh infra..."

    istioctl install -y --verify -f "$repo_dir"/deployment/mesh-infra/istio/profile.yaml
    kubectl label namespace default istio-injection=enabled || error_exit "Failed to label default namespace for istio injection."
}

setup_storage() {
    log "Creating storage directories..."
    sudo mkdir -p /mnt/data || error_exit "Failed to create /mnt/data directory."
    sudo chmod 777 /mnt/data || error_exit "Failed to set permissions on /mnt/data."
    sudo mkdir -p /mnt/data/d{1..10} || error_exit "Failed to create /mnt/data directories."

    log "Setting up Persistent Volumes..."
    pv_tool="$repo_dir/utils/create-local-pv.sh"
    bash "$pv_tool" /mnt/data/d1 -s 20Gi -n local-pv-1 || error_exit "Failed to create Persistent Volume for d1."
    for i in {2..10}; do
        bash "$pv_tool" "/mnt/data/d$i" -s 10Gi -n "local-pv-$i" || error_exit "Failed to create Persistent Volume for d$i."
    done

    log "Local-static-provisioner storage setup completed."
}

setup_argocd() {
    argocd_dir="$repo_dir/deployment/mesh-infra/argocd"
    log "Setting up ArgoCD from $argocd_dir..."

    # First add the ArgoCD namespace
    kubectl apply -f "$argocd_dir"/namespace.yaml #>/dev/null || error_exit "Failed to create ArgoCD namespace."

    # Create secrets
    nix run .#node-config -- -microk8s basicnode-secrets #|| error_exit "Failed to create ArgoCD secrets."

    # Apply first and ignore errors
    #kubectl apply -k "$argocd_dir" #>/dev/null || log "Initial ArgoCD apply encountered errors, proceeding..."
    #kubectl apply -k "$argocd_dir" #>/dev/null || error_exit "Failed to apply ArgoCD configuration."
    
    #kustomize build `echo "$repo_dir""/deployment/mesh-infra/argocd"` | kubectl apply -f -

    #kustomize build `echo "$repo_dir""/deployment/mesh-infra/argocd"` | kubectl apply -f -
    kustomize build "$repo_dir"/deployment/mesh-infra/argocd | kubectl apply -f -
    
    kustomize build "$repo_dir"/deployment/mesh-infra/argocd | kubectl apply -f -
}

main "$@"

