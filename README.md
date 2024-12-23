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

We recommend to deploy a TEADAL node on a machine with 8 cores, 32 GB memory, 100GB storage. Depending on the TEADAL tools installed less or more than these resources could be required.

To deploy a TEADAL node follow this [Teadal Node Installation guide](docs/InstallTeadalNode.md)

To add a TEADAL node follow [Teadal Tool Installation guide](docs/InstallTeadalTools.md)