# Teadal tools installation

This document provides a guide to install the tools developed in the TEADAL project. These tools requires a TEADAL node up and running (see [Teadal Node installation guide](QuickStart.md)) 

At this time, the following tools are available:
- [Advocate](#advocate)
- [Catalogue](#catalogue)
- [Policy manager](#policy)


If any dependencies are indicated, then it is required to configure the ArgoCD project in your Teadal Node to add the required tool.

## Advocate <a name="advocate"/>

#### Dependencies

- Jaeger

See the related [page](InstallAddons.md) to know how to install the dependencies

#### Preliminary steps

Before Advocate will work you will need to create the related namespace
```bash
kubectl create namespace trust-plane
```
Then it is required to configure all needed secrets, variables for Advocate blockchain such as wallet private key, VM key and Ethereum Remote Procedure Call (RPC) Address. For that run this command:
```bash
node.config --microk8s advocate
```
Now you can enter the required values. For the question about the "ADVOCATE_ETH_POA" , enter "1" as value.


#### Tool deployment

Be sure that, **on the repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/plat-infra-services/kustomization.yaml) used by argocd has the line ``- advocate`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
- advocate
```

After few minutes, ArgoCD will realizes the update and starts deploying the related pods.


#### Checking installation

Check pods that are in Trust-plane namespace:

```bash
kubectl get pods -n trust-plane
```

![screenshot](./images/trust-plane-namespace-podes.png)

Check the Advocate pod log to make sure that it is up and running:

```bash
kubectl logs <advocate-pod-name> -n trust-plane
```
![screenshot](./images/advocate-pod-log.png)


## Catalogue deployment <a name="catalog"/>

#### Dependencies

TBD

#### Preliminary steps

TBD

#### Tool deployment

TBD

#### Checking installation

TBD

## Policy manager <a name="policy"/>

#### Dependencies

TBD

#### Preliminary steps

TBD

#### Tool deployment

TBD

#### Checking installation

TBD