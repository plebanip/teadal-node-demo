### Setup the enviroment

First off, you should install Nix and enable the Flakes extension.

For that connect to the machine you want to install the Teadal node to and then follow the guide below. 

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

If your VM has slow I/O Disk (like in case of POLIMI testbed), it is suggested to remove also the high availability adds-on. To do it, the following command is required

```
microk8s disable ha-cluster --force
```

Finally bolt on DNS and local storage

```
microk8s enable dns
microk8s enable hostpath-storage
```

Wait until all the above extras show in the "enabled" list and the removed ha-cluster is in the "disabled" list

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
Something like this should appear

![screenshot](./images/microk8s-1.png)

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

We'll start off with microk8s `hostpath-storage` for now since we've only got one
node in the cluster. Later on, when we add more nodes, we'll switch
over to distributed storage backed by local disks on each node. (We
set up DirectPV for that, but we could also use Longhorn or something
else.)


[comment]: # (We'll create 4 PVs of 5GB each and 1 PV of 20GB. Ideally they should be backed by disk partitions, but we'll cheat a bit and create dirs straight into the `/mnt` directory. For the record, here's the proper way of doing [this sort of thing][proper-ls]. Anyhoo, let's go on with creating the dirs. SSH into the target node, then)

[comment]: # (```bash)
[comment]: # (sudo mkdir -p /data/d{1..5})
[comment]: # (sudo chmod -R 777 /data)
[comment]: # (```)

Make sure the `hostpath-storage` addon is enabled on microk8s by checking the presence of this add-on the in the 'enabled' list

```
microk8s status
```

If every PersistentVolumeClaim has it's StorageClass field set to `microk8s-hostpath`,
the microk8s addon will helpfully provision all PersistentVolumes necessary, so 
no further action is necessary here. By default, the code should be already configured to be compliant with this approach.


#### K8s secrets

```bash
kubectl apply -f mesh-infra/argocd/namespace.yaml
```

Edit the K8s Secret templates in `mesh-infra/security/secrets` to
enter the passwords you'd like to use. 

Then install them in the cluster.

```bash
kustomize build mesh-infra/security/secrets | kubectl apply -f -
```

#### Istio

Don't install Istio as a MicroK8s add-on, since MicroK8s will install an old version! For this reason, it is required to follow the following procedure

Deploy Istio to the cluster using our own profile

```bash
istioctl install -y --verify -f mesh-infra/istio/profile.yaml
```

For now platform infra services (e.g. DBs) as well as app services
(e.g. file transfer UI) sit in K8s' `default` namespace, so tell Istio
to auto-magically add an Envoy sidecar to each service deployed to
that namespace

```bash
kubectl label namespace default istio-injection=enabled
```

Notice that you can actually be selective about which services get
an Envoy sidecar, but for now we'll just apply a blanket policy to
keep things simple.

A final check to see if istio is deployed in k8s
```
kubectl get pod -A
```

#### Argo CD

Argo CD is our declarative continuous delivery engine. Except for
the things listed in this bootstrap procedure, we declare the cluster
state with YAML files that we keep in the `deployment` dir within
our GitHub repo. Argo CD takes care of reconciling the current cluster
state with what we declared in the repo.

For that to happen, we've got to deploy Argo CD and tell it to use
the YAML in our repo to populate the cluster. Our repo also contains
the instructions for Argo CD to manage its own deployment state as
well as the rest of the Teadal platform—I know, it sounds like a dog
chasing its own tail, but it works. So we can just build the YAML to
deploy Argo CD and connect it to our repo like this

```bash
kustomize build mesh-infra/argocd | kubectl apply -f -
```

After deploying itself to the cluster, Argo CD will populate it with
all the K8s resources we declared in our repo and so slowly the Teadal
platform instance will come into its own. This will take some time.
Go for coffee (actually also for lunch or dinner as it could take more than 1 hour).

> Note
>* Argo CD project errors. If you see a message like the one below in
  the output, rerun the above command again — see [#42][boot.argo-app-issue]
  about it.
>> unable to recognize "STDIN": no matches for kind "AppProject" in version "argoproj.io/v1alpha1"

Notice that Argo CD creates an initial secret with an admin user of
`admin` and randomly generated password on the first deployment. To
grab that password, run

```bash 
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d && echo
```

You can use it if you get in trouble during the bootstrap procedure,
but keeping it around is like an accident waiting to happen. So you
should definitely zap it as soon as you've managed to log into Argo
CD with the password you entered in our secret. To do that, just

```bash
kubectl -n argocd delete secret argocd-initial-admin-secret
```

If this secret is not returned, it is needed to generate a new admin password and configure the node with this new password (see here for more information https://github.com/argoproj/argo-cd/blob/master/docs/faq.md)). To generate the password

```
argocd account bcrypt --password <YOUR-PASSWORD-HERE>
```

Copy the generate password and attach it to the following command

```
# bcrypt(password)=$2a$10$rRyBsGSHK6.uc8fntPwVIuLVHgsAhAX7TcdrqW/RADU0uh7CaChLa
kubectl -n argocd patch secret argocd-secret \
  -p '{"stringData": {
    "admin.password": "$2a$10$rRyBsGSHK6.uc8fntPwVIuLVHgsAhAX7TcdrqW/RADU0uh7CaChLa",
    "admin.passwordMtime": "'$(date +%FT%T%Z)'"
  }}'
```

Again, check to see if argocs is deployed in k8s as well
```
kubectl get pod -A
```

After sometime this command returns the complete set of pods up and running. Good news! Your node is working!!!

### Checking the installation

You can now test the installation to see if everything is working

#### Argo CD

First off, browse to the Argo CD web app for a bird's eye view of
what the GitOps pipeline has deployed. Istio routes HTTP traffic
to port 80 and path `/argocd` to the Argo CD server inside the mesh.
So the web app is available at

- http://localhost/argocd/

Log in with the username and password you entered in the Argo CD K8s
secret. K8s resources are grouped into apps and, in turn, apps into
three projects (mesh-infra, plat-infra-services, plat-app-services)
to reflect the layout of the `deployment` dir in the Git repo.


### Istio

Istio got configured with a few useful add-ons, HTTP and TCP routing,
plus some settings to make it easier to intercept and debug raw network
packets. Any service deployed to the default namespace automatically
gets an Istio side car.

```bash
$ kubectl get pod
```

You should see Keycloak, Postgres and HttpBin entries, all having a
pod count of 2: one pod is the service proper, the other's the Istio
sidecar.

To see some mesh action, use `istioctl` to connect Kiali (one of the
installed add-ons) to your browser

```bash
$ istioctl d kiali
```

navigate to the HttpBin workload and then generate some traffic by
hitting the HttpBin SVG image endpoint many times in a row until
you see the Kiali graph

- http://localhost/httpbin/image/svg

(This works because the GitOps pipeline defined an HTTP route through
port `80` and path `/httpbin`.)
Have a look at the content of the other tabs in Kiali, especially
the traces to see Jaeger, another installed add-on, at work. Then
take a peek at the various performance dashboards

```bash
$ istioctl d grafana
```

Read more about the add-ons we bundled with Istio over here

- https://github.com/istio/istio/tree/1.18.0/samples/addons

Oh, truth be told, we also have SkyWalking. But that's not deployed
yet because we still have some bugs to squash.


### Storage

Three of the PVs should be bound. One for the Postgres DB, another
for Keycloak's DB and the last one for the Teadal MinIO tenant.

```bash
$ kubectl get pv
$ kubectl get pvc -A
```

Log into the MinIO console. Run

```bash
$ kubectl minio proxy
```

then browse to http://localhost:9090 and enter the JWT token printed
on the terminal. You should see the Teadal tenant the GitOps pipeline
created. Istio routes the tenant's S3 service through port `80` and
path `/minio`. So you should be able to hit the service from your box
like this

```bash
$ curl -v http://localhost/minio
```

You should get a fat `403` response. Access is denied without valid
creds. How rude.


### Security

Keycloak is at http://localhost/keycloak, courtesy of Istio routing.
Navigate to the admin console and log in with the username and password
you set in the Keycloak K8s secret.

Speaking of secrets, we've got Reloader to watch for secret changes
and bounce affected pods. So secrets always stay fresh. Have a look
at the K8s logs to see what's going on under the bonnet.

After logging onto Keycloak you should check there's a Teadal realm
with two users, `jeejee@teadal.eu` and `sebs@teadal.eu`, both having
log-in enabled and a password of `abc123`. The GitOps pipeline creates
and pre-populates the Teadal realm automatically.

Now you should check out how we secure data products. Our setup ain't
exactly straightforward, so to make sense of the examples below you
should probably first read about [our security architecture][sec],
at least the conceptual model section.

We'll use HttbBin to simulate a data product. There's a [policy][httpbin-rbac]
that defines two roles:
- *Product owner*. The owner may do any kind of HTTP request to URLs
   starting with `/httpbin/anything/`.
- *Product consumer*. On the other hand, the consumer is only allowed
  to read (`GET`) URLs starting with `/httpbin/anything/` or the (exact)
  URL `/httpbin/get`.

`jeejee@teadal.eu` is both a product owner and consumer, whereas
`sebs@teadal.eu` is just a consumer. To interact with HttpBin, both
users need to get a JWT from Keycloak and attach it to service requests
since the policy doesn't allow anonymous requests to the above URLs.
In fact, if you try e.g.

```bash
$ curl -i -X GET localhost/httpbin/anything/do
$ curl -i -X GET localhost/httpbin/get
```

you should get back a fat `403` in both cases. So let's get a JWT
for `jeejee@teadal.eu`. We'll store it in a env var so we can use
it later. The command below should do the trick. (If you've changed
the user's password in Keycloak, replace `abc123` with the new one.)

```bash
$ export jeejees_token=$(\
    curl -s \
      http://localhost/keycloak/realms/teadal/protocol/openid-connect/token \
      -d 'grant_type=password' -d 'client_id=admin-cli' \
      -d 'username=jeejee@teadal.eu' -d 'password=abc123' | jq -r '.access_token')
```

And, as we're at it, let's get a JWT for `sebs@teadal.eu` too.

```bash
$ export sebs_token=$(\
    curl -s \
      http://localhost/keycloak/realms/teadal/protocol/openid-connect/token \
      -d 'grant_type=password' -d 'client_id=admin-cli' \
      -d 'username=sebs@teadal.eu' -d 'password=abc123' | jq -r '.access_token')
```

Again, if you've changed `sebs@teadal.eu`'s password, update the
command above accordingly. Also keep in mind these tokens are quite
short-lived (about 4 mins) so if you take too long to go through the
examples below, you'll have to get fresh tokens again.

Both product owner and consumer are allowed to read a URL path like
`/httpbin/anything/do`. So both users, `jeejee@teadal.eu` (owner)
and `sebs@teadal.eu` (consumer), should be able to do a GET and get
back (pun intended) a `200`, provided we attach their respective JWT
to the each request:

```bash
$ curl -i -X GET localhost/httpbin/anything/do \
       -H "Authorization: Bearer ${jeejees_token}"
$ curl -i -X GET localhost/httpbin/anything/do \
       -H "Authorization: Bearer ${sebs_token}"
```

But, as a product owner, `jeejee@teadal.eu` should be able to do
anything he fancies to the above path, like `DELETE`, whereas
`sebs@teadal.eu`, as a consumer, should not.

```bash
$ curl -i -X DELETE localhost/httpbin/anything/do \
       -H "Authorization: Bearer ${jeejees_token}"
$ curl -i -X DELETE localhost/httpbin/anything/do \
       -H "Authorization: Bearer ${sebs_token}"
```

You should see a `200` response for the first request, but a `403`
for the second. Finally, since both users are product consumers,
they should both allowed to `GET /httpbin/get`

```bash
$ curl -i -X GET localhost/httpbin/get \
       -H "Authorization: Bearer ${jeejees_token}"
$ curl -i -X GET localhost/httpbin/get \
       -H "Authorization: Bearer ${sebs_token}"
```

You should see a `200` response in both cases. That just about wraps
it up for the security show.


### DBs

At the moment we only have Postgres. Istio routes incoming TCP traffic
from port `5432` to the Postgres server. Here's an easy way to get
into the DB with `psql` from your local machine.

```
$ nix-shell -p postgresql_15
$ psql postgres://postgres:abc123@localhost
```

Replace `abc123` with the password you entered in the Postgres K8s
secret.

[httpbin-rbac]: ../deployment/mesh-infra/security/opa/rego/httpbin/rbacdb.rego
[sec]: ./sec-design/README.md
