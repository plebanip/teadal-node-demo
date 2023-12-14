### Setup the enviroment

First off, you should install Nix and enable the Flakes extension

```
sh <(curl -L https://nixos.org/nix/install) --daemon
mkdir -p ~/.config/nix
echo 'experimental-features = nix-command flakes' >> ~/.config/nix/nix.conf
```

clone the repo
```
git clone https://gitlab.teadal.ubiwhere.com/teadal-pilots/<name of pilot>/<name of pilot>.git
```

run the nix shell under the just cloned repo
```
cd <clonerepo dir>/nix
nix shell
```

check if it worked by checking the ArgoCD version 
```
argocd version --client --short
```
it should return something like ``argocd: v2.7.6``

Now all the command must be executed inside the Nix shell

### Install MicroK8S

We'll use MicroK8s as a cluster manager and orchestration. Install MicroK8s (upstream Kubernetes 1.27)

```
sudo snap install microk8s --classic --channel=1.27/stable
```

Add yourself to the MicroK8s group to avoid having to sudo every time your run a microk8s command

```
sudo usermod -a -G microk8s $(whoami)
newgrp microk8s
```

and then wait until MicroK8s is up and running
```
microk8s status --wait-ready
```

Finally bolt on DNS and local storage

```
microk8s enable dns
microk8s enable hostpath-storage
```

Wait until all the above extras show in the "enabled" list

```
microk8s status
```
