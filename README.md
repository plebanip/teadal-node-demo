![alt text](docs/images/teadal-logo.jpg)

# TEADAL node

A TEADAL node is the cornerstone platform developed in the in TEADAL project to host a toolset able to simplify the data sharing for analytical purposes

The TEADAL node is built on a K8s cluster (now we have tested with MicroK8s) and requires the following tools:

* ArgoCD to automate the deployment
* Istio control plane
* Minio to store system files and also available as data lake storage
* Keycloak for Identity and Authentication
* Jaeger for advance tracing used to monitor the activities
* OPA as policy manager

In addition to these tools, TEADAL project is providing advanced tools to enable data sharing among TEADAL nodes:

* Advocate
* Catalogue
* Policy manager
* Pipeline generator



## Quick installation
1) Fork this repo 
2) Download the [installation file](utils/install-teadal.sh) on your machine [*]
3) Make the file installation file executable `sudo chmod 777 install-teadal.sh`
4) Run the installation file `./install-teadal.sh -u <git repo url> [-b <branch name> -d <download dir>]`
where:
    - `<git repo url>` is the URL of the forked repo
    - `<branch name>` the branch of the repo, (default: `HEAD`)
    - `<download dir>` directory where the repo is cloned (default: `./teadal.node`)  

[*] We recommend to deploy a TEADAL node on a machine with 8 cores, 32 GB memory, 100GB storage. Depending on the TEADAL tools installed less or more than these resources could be required.

### Detailed installation procedure
Additional information about the installation process is available in the [Teadal Node Installation guide](docs/InstallTeadalNode.md)

### TEADAL node configuration
To add a TEADAL tools follow [Teadal Tool Installation guide](docs/InstallTeadalTools.md)

## Adding an FDP

A Federated Data Product is defined by its descriptor which has the following format

```yaml
fdp-descriptor:
  name: <name_of_the_FDP>
  version: 1.0.0
  description: >
    FDP descriptor used to test

  deployment: microk8s_deployment.yaml      
  resources:
  - additional.yaml      
```
The `microk8s_deployment` file follows the format required by microk8s. An example is [here](/deployment/plat-app-services/httpbin/base.yaml). In case the deployment file requires additional files they can be defined in the `resources` list 

Having the descriptor execute the following commands:

```bash
cd <download dir>/nix/pkgs/fdp-manager
nix run .#fdp-manager deploy <descriptorfile>
```

For testing, under the `example` folder there is the `fdp-sync-dummy` FDP that can be used.

```bash
nix run .#fdp-manager deploy ~/teadal.node/examples/fdp-sync-dummy/descriptor.yaml
```
Note! To make the deployment effective, the changes in the repository must be committed

For additional information about the FDP-manager refer to the [related project](https://gitlab.teadal.ubiwhere.com/teadal-tech/fdp-manager)

### Detailed FDP deployment

Additional information about the deployment is provided [here](docs/AddingPilot-FDP.md)

