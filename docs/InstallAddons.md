# Add-ons installation

This document provides a guide to install common tools that are pre-requisites for some TEADAL tools. Most of the time, recipes for the deployment of these tools are already defined in the repo, thus it is an easy task. These add-ons are grouped in two classes:

- [Generic tools](#generic)
- [Istio-related tools](#istio-related)

Information about how to unistall these tools can be found [here](UninstallTeadalElements.md).

## Generic add-ons installation <a name="generic"/>

The generic tools for which a kustomization is already included are:

- [Kepler](#kepler)

**Note!** For this Istio-related add-ons, once changed the kustomization file, remember to push the modification to the git repo to inform ArgoCD about the changes.


### Kepler <a name="kepler"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) used by argocd has the line ``- kepler`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- kepler
...
```



## Istio add-ons installation <a name="istio-related"/>

The Istio-relaated tools for which a kustomization is already included are:

- [Jaeger](#jaeger)
- [Kiali](#kiali)
- [Grafana](#grafana)
- [Airflow](#airflow)
- [Prometheus (with Thanos)](#prometheus)

**Note!** For these Istio-related add-ons, once changed the kustomization file, remember to:
- push the modification to keep the repo aligned with the new configuration
- update istio by executing the command ``kustomize build deployment/mesh-infra/istio | kubectl apply -f -``

### Jaeger <a name="jaeger"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) used by argocd has the line ``- jaeger`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- jaeger
...
```

### Kiali <a name="kiali"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) used by argocd has the line ``- kiali`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- kiali
...
```

### Grafana <a name="grafana"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) used by argocd has the line ``- grafana`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- grafana
...
```

### Airflow <a name="airflow"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/plat-infra-services/kustomization.yaml) used by argocd has the line ``- airflow`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- airflow
...
```

### Prometheus <a name="prometheus"/>


Prometheus deployment is already configured to be installed in a plain mode or equipped with Thanos. 

#### Prometheus plain

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) has the line ``- prometheus`` uncommented and the line ``- thanos`` commented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- prometheus
#- thanos
...
```

Then, the [kustomization file](../deployment/mesh-infra/istio/prometheus/kustomization.yaml) has the line ``- prometheus`` uncommented, while the lines ``- prometheus-thanos`` and ``- node-exporter-daemonset.yaml`` commented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- prometheus
#- prometheus-thanos
#- node-exporter-daemonset.yaml
...
```


#### Prometheus with Thanos

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/mesh-infra/kustomization.yaml) has both the lines ``- prometheus`` and ``- thanos`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- prometheus
- thanos
...
```

Then, the [kustomization file](../deployment/mesh-infra/istio/prometheus/kustomization.yaml) has the lines ``- prometheus-thanos`` and ``- node-exporter-daemonset.yaml`` uncommented, while the line ``- prometheus`` commented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
#- prometheus
- prometheus-thanos
- node-exporter-daemonset.yaml
...
```

In the Prometheus+Thanos configuration, a Minio instance must be present in the cluster.