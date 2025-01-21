#!/bin/bash

# Function to display usage information
usage() {
    echo "Usage: $0 [OPTIONS] repoURL"
    echo "Mandatory parameters:"
    echo "  -d <repo_dir>     Specify the directory with the repo clone"
    echo "  -r <repo_url>     Specify the repoURL"
    echo "Options:"
    echo "  -b <branch>       Specify a branch"
    echo "  -h                Display this help message"
    exit 1
}

# check pre-requisites
if [ -z "$HOSTNAME" ]; then
   echo "$HOSTNAME variable not defined"
   exit1
else 
   echo "installing Teadal node in $HOSTNAME"
fi


# Initialize variables
repo_dir=""
repo_url=""
branch=""
hostname_dir=""

# Parse command-line options
while getopts ":d:r:b:h" opt; do
    case "$opt" in
        d)
            repo_dir="$OPTARG"
            ;;
        r)
            repo_url="$OPTARG"
            ;;
        b)
            branch="$OPTARG"
            ;;
        h)
            usage
            ;;
        ?)
            echo "Invalid option: -$OPTARG" >&2
            usage
            ;;
    esac
done



# Check if the mandatory parameter (-d) is provided
if [ -z "$repo_dir" ]; then
    echo "Error: The -d <repo_dir> parameter is mandatory." >&2
    usage
fi

# Check if the mandatory parameter (-r) is provided
if [ -z "$repo_url" ]; then
    echo "Error: The -r <repo_url> parameter is mandatory." >&2
    usage
fi

echo "### installing microk8s ###"

echo "sudo snap install microk8s --classic --channel=1.27/stable"
sudo snap install microk8s --classic --channel=1.27/stable

echo "sudo usermod -a -G microk8s $(whoami)"
sudo usermod -a -G microk8s $(whoami) 2> error.log

sudo newgrp microk8s << MYGRP
echo "newgrp microk8s"
MYGRP

echo "microk8s status --wait-ready"
microk8s status --wait-ready

echo "### microk8s installed ###"

echo "### configuring microk8s ###"

microk8s disable ha-cluster --force
microk8s enable dns
microk8s status

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
    echo "$replacement" >> "$file"    
fi

microk8s stop
microk8s start

export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
microk8s config > ~/.kube/config
kubectl get pod -A
echo "### microk8s configured ###"


echo "setting up microk8s storage"

sudo mkdir -p /mnt/data/d{1..10}
sudo chmod -R 777 /mnt/data
node.config -microk8s pv 1:20 8:10
hostname_dir=`echo "$HOSTNAME" | tr '[:upper:]' '[:lower:]'`
echo "pippo"
echo "$hostname_dir"
mv "$hostname_dir" "$repo_dir"/deployment/mesh-infra/storage/pv/local/

# change the kustomizefile for storage ports
file="$repo_dir""/deployment/mesh-infra/storage/pv/local/kustomization.yaml"
kustomizationfile_dir="$repo_dir""/deployment/mesh-infra/storage/pv/local/"
substring="\- <HOST_NAME>"
replacement=`echo "-" $hostname_dir`

# Check if the file exists
if [ ! -f "$file" ]; then
    echo "Error: File '$file' does not exist."
    exit 1
fi

# Check if the file contains a line with the given substring
if grep -q "$substring" "$file"; then
    sed -i.bak "s/^$substring.*/$replacement/" "$file" && rm "$file".bak
else
    substring=`echo "\-" "$hostname_dir"`
    if grep -q "$substring" "$file"; then
        echo "folder already included in the  kustomizationfile"
    else
        echo "$replacement" >> "$file"
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
replacement=`echo "    repoURL:" "$repo_url"`

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

kustomize build `echo "$repo_dir""/deployment/mesh-infra/argocd"` | kubectl apply -f -

#try twice
kustomize build `echo "$repo_dir""/deployment/mesh-infra/argocd"` | kubectl apply -f -

kubectl get pod -A

node.config -microk8s basicnode-secrets

echo "ArgoCD installed"

echo "should be done"
