#!/bin/bash
set -euo pipefail

# Function to display usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Mandatory options:"
    echo "  -d <repo_dir>     Specify the directory with the repo clone"
    echo "  -r <repo_url>     Specify the repoURL"
    echo "Options:"
    echo "  -b <branch>       Specify a branch"
    echo "  -h                Display this help message"
}

main() {
    parse_options "$@"
    log "Script started with options:\n\trepo_dir=$repo_dir\n\trepo_url=$repo_url\n\tbranch=$branch"
    setup_microk8s
    # exit 0 # TODO: comment before commit until fully tested
}

### Utilities scripts
# Colors and logging functions
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

TEADAL_LOG_DIR="${TEADAL_LOG_DIR:-/tmp}"
logfile="$TEADAL_LOG_DIR/install-teadal.log"

log() { echo "${GREEN}[INFO]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a $logfile; }
error() { echo "${RED}[ERROR]${NC}$(date +'%Y-%m-%d %H:%M:%S') - $1" | tee -a $logfile >&2; }

# Global variables
repo_dir="$(pwd)" # Directory with the repo clone
repo_url="$(git config --get remote.origin.url 2>/dev/null || echo '')"
# Url of the repo
branch=""       # Branch of the repo
hostname_dir="" # Directory with generated storage pv

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
        sudo snap install microk8s --classic --channel=1.27/stable || error "Failed to install microk8s."
    fi

    # Setup permissions
    sudo usermod -a -G microk8s $USER
    mkdir -p ~/.kube
    chmod 0700 ~/.kube
    # Setup addons
    log "Waiting for microk8s to be ready..."
    sudo microk8s status --wait-ready &>/dev/null || error "microk8s is not ready."
    log "Enabling microk8s addons (it may take a while)..."
    sudo microk8s enable dns &>/dev/null || error "Failed to enable dns."
    sudo microk8s disable ha-cluster --force &>/dev/null || error "Failed to disable ha-cluster."
    sudo microk8s config >~/.kube/config
    export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
    log "Waiting for microk8s to be ready after enabling addons..."
    sudo microk8s status --wait-ready
    log "microk8s is ready."
}

main "$@"

# change the kube-apiserver ports
file="/var/snap/microk8s/current/args/kube-apiserver"
substring="\-\-service-node-port-range"
replacement="--service-node-port-range=1-65535"

# Check if the file exists
if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

# Check if the file contains a line with the given substring
if grep -q "$substring" "$file"; then
    sed -i.bak "s/^$substring.*/$replacement/" "$file" && rm "$file".bak
else
    echo "$replacement" >>"$file"
fi

microk8s stop
microk8s start

export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
microk8s config >~/.kube/config
kubectl get pod -A
echo "### microk8s configured ###"

echo "setting up microk8s storage"

sudo mkdir -p /mnt/data/d{1..10}
sudo chmod -R 777 /mnt/data
node.config -microk8s pv 1:20 8:10
hostname_dir=$(echo "$HOSTNAME" | tr '[:upper:]' '[:lower:]')
echo "pippo"
echo "$hostname_dir"
mv "$hostname_dir" "$repo_dir"/deployment/mesh-infra/storage/pv/local/

# change the kustomizefile for storage ports
file="$repo_dir""/deployment/mesh-infra/storage/pv/local/kustomization.yaml"
kustomizationfile_dir="$repo_dir""/deployment/mesh-infra/storage/pv/local/"
substring="\- <HOST_NAME>"
replacement=$(echo "-" $hostname_dir)

# Check if the file exists
if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

# Check if the file contains a line with the given substring
if grep -q "$substring" "$file"; then
    sed -i.bak "s/^$substring.*/$replacement/" "$file" && rm "$file".bak
else
    substring=$(echo "\-" "$hostname_dir")
    if grep -q "$substring" "$file"; then
        echo "folder already included in the  kustomizationfile"
    else
        echo "$replacement" >>"$file"
    fi
fi

kustomize build "$kustomizationfile_dir" | kubectl apply -f -

kubectl get pv

echo "microk8s storage set"

echo "installing istio"

istioctl install -y --verify -f "$repo_dir"/deployment/mesh-infra/istio/profile.yaml
kubectl label namespace default istio-injection=enabled

kubectl get pod -A

echo "istio installed"

echo "installing ArgoCD"

# change the kustomizefile for argocd repo
file="$repo_dir""/deployment/mesh-infra/argocd/projects/base/app.yaml"
substring="repoURL"
replacement=$(echo "    repoURL:" "$repo_url")

# Check if the file exists
if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

# Check if the file contains a line with the given substring
if grep -q "$substring" "$file"; then
    sed -i.bak "s/^$substring.*/$replacement/" "$file" && rm "$file".bak
    echo "$file" " updated with " "$replacement"
else
    echo "Error the repoURL field does not exist"
fi

if [ -z "$branch"]; then
    # change the kustomizefile for argocd repo
    file="$repo_dir""/deployment/mesh-infra/argocd/projects/base/app.yaml"
    substring="targetRevision"
    replacement="targetRevision: $branch"

    # Check if the file contains a line with the given substring
    if grep -q "$substring" "$file"; then
        sed -i.bak "s/^$substring.*/$replacement/" "$file" && rm "$file".bak
    else
        echo "ArgoCD customisation file must have targetRevision field"
    fi
fi

kustomize build $(echo "$repo_dir""/deployment/mesh-infra/argocd") | kubectl apply -f -

#try twice
kustomize build $(echo "$repo_dir""/deployment/mesh-infra/argocd") | kubectl apply -f -

kubectl get pod -A

node.config -microk8s basicnode-secrets

echo "ArgoCD installed"

echo "should be done"
