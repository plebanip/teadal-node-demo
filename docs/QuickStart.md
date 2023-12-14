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

Now we've got to broaden MicroK8s node port range. This is to make sure it'll be able to expose any K8s node port we're going to use.
```
nano /var/snap/microk8s/current/args/kube-apiserver
```

and add this line somewhere in the file
```
--service-node-port-range=1-65535
```

Then restart microk8s
```
microk8s stop
microk8s start
```


> Notes
> * Istio. Don't install Istio as a MicroK8s add-on, since MicroK8s will > install an old version!
> * Storage. MicroK8s comes with its own storage provider (microk8s.io/>hostpath) which the storage add-on enables as well as creating a default K8s storage class called microk8s-hostpath.


Set up the KUBECONFIG variable to make kubectl accessible
```
export KUBECONFIG=/var/snap/microk8s/current/credentials/client.config
```
> Note to make the k8s accessible from outside of the VM
> Copy out the K8s admin creds
> ```
> cat /var/snap/microk8s/current/credentials/client.config
> ```
> save them to a local file outside the VM and replace the IP address of the server URL with that of your Multipass VM, e.g.
server: https://192.168.64.28:16443
> Run the following command outside the VM to grab the IP address
>
>Finally, export KUBECONFIG so kubectl, istioctl and friends know where the cluster is
> ```
> export KUBECONFIG=/path/to/your/copy/of/client.config.  
> ```

Check the status of k8s
```
kubectl get pod -A
```
### Setup the network

The mesh we're going to roll out needs to be connected to some ports
on the external network. Clients on the external network hit port `80`
to access HTTP services. The Istio gateway uses a K8s node port to
accept incoming traffic on port `80` and route it to the destination
service inside the mesh. The Istio gateway also has a `5432` node port
to let external clients interact with the Postgres DB inside the mesh.
Additionally, the node port `3810` is configured on the Istio gateway 
to route traffic to the kubeflow UI service.
Finally admins will want to SSH into cluster nodes so port `22` should
be open too as well as port `6443` which is the K8s API endpoint admin
tools like `kubectl` should connect to.

How you actually make these ports available to processes running
outside the mesh really depends on your setup. In the most trivial
case where your cluster is made up by a single node and that node
is directly connected to the Internet, all you need to do is open
those ports in the firewall, if you have a one, or do nothing if
there's no firewall. In a public cloud scenario, e.g. AWS, you
typically have an admin console that lets you easily make ports
available to clients out in the interwebs.

### Setup the mesh

First of all made $dir$/deployment/ your current dir

#### K8s storage

We'll start off with local storage for now since we've only got one
node in the cluster. Later on, when we add more nodes, we'll switch
over to distributed storage backed by local disks on each node. (We
set up DirectPV for that, but we could also use Longhorn or something
else.)

We'll create 4 PVs of 5GB each and 1 PV of 20GB. Ideally they should be backed by
disk partitions, but we'll cheat a bit and create dirs straight into
the `/mnt` directory. (For the record, here's the proper way of
doing [this sort of thing][proper-ls].) Anyhoo, let's go on with
creating the dirs. SSH into the target node, then

```bash
$ sudo mkdir -p /data/d{1..5}
$ sudo chmod -R 777 /data
```

Now get back to your Teadal repo on your local machine and run

```bash
$ kustomize build mesh-infra/storage/pv/local/devm/ | kubectl apply -f -
```


#### K8s secrets

```bash
$ kubectl apply -f mesh-infra/argocd/namespace.yaml
```

Edit the K8s Secret templates in `mesh-infra/security/secrets` to
enter the passwords you'd like to use. Then install them in the cluster

```bash
$ kustomize build mesh-infra/security/secrets | kubectl apply -f -
```

