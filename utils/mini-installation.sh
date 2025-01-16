sudo snap install microk8s --classic --channel=1.27/stable
microk8s status --wait-ready
microk8s disable ha-cluster --force
microk8s enable dns
microk8s status
echo "--service-node-port-range=1-65535" >> /var/snap/microk8s/current/args/kube-apiserver
kustomize build mesh-infra/storage/pv/local/ | kubectl apply -f -
kustomize build mesh-infra/storage/pvc/ | kubectl apply -f -
microk8s stop
microk8s start
microk8s config > ~/.kube/config
export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
istioctl install -y --verify -f mesh-infra/istio/profile.yaml
kubectl label namespace default istio-injection=enabled
kubectl get pod -A
