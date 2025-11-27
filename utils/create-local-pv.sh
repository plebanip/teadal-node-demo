#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: $0 <hostname_dir> [OPTIONS]"
    echo "Mandatory arguments:"
    echo "  <hostname_dir>      Specify the mount directory of local PVs (required)"
    echo "Options:"
    echo "  -c --class <storage_class>   Specify the storage class (default: local-storage)"
    echo "  --dry-run                    Do not apply the generated PV to the cluster"
    echo "  -s --size <size>             Specify the size of the PV (default: 100Gi)"
    echo "  -n --name <name>             Specify the name of the PV (default: local-pv-<hostname>)"
    exit 1
}

hostname_dir=""
storage_class="local-storage"
no_apply=false
size="100Gi"

if [ $# -lt 1 ]; then
    usage
fi
if [ "$1" == "-h" ] || [ "$1" == "--help" ]; then
    usage
fi
hostname_dir="$1"
shift

while [[ $# -gt 0 ]]; do
    case $1 in
    -c | --class)
        storage_class="$2"
        shift 2
        ;;
    --dry-run)
        no_apply=true
        shift
        ;;
    -s | --size)
        size="$2"
        shift 2
        ;;
    -n | --name)
        name="$2"
        shift 2
        ;;
    -*)
        echo "Unknown option: $1"
        usage
        ;;
    *)
        echo "Unexpected argument: $1"
        usage
        ;;
    esac
done

if [ -z "$hostname_dir" ]; then
    usage
fi

if [ -z "${name:-}" ]; then
    name="local-pv-$(basename "$hostname_dir")"
fi

if [ ! -d "$hostname_dir" ]; then
    echo "Error: Directory $hostname_dir does not exist."
    exit 1
fi

pv_yaml="cat <<EOF
apiVersion: v1
kind: PersistentVolume
metadata:
  name: $name
spec:
  capacity:
    storage: $size
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: $storage_class
  local:
    path: $hostname_dir
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - $(hostname)
EOF"

if [ "$no_apply" = true ]; then
    eval "$pv_yaml"
else
    echo "Creating Persistent Volume for directory $hostname_dir with storage class $storage_class..."
    eval "$pv_yaml" | kubectl apply -f -
    echo "Persistent Volume created."
fi
