# Add-ons installation

This document provides a guide to install common tools that are pre-requisites for some TEADAL tools. Most of the time, recipes for the deployment of these tools are already defined in the repo, thus it is an eaasy task

- [Jaeger](#jaeger)
- [Kepler](#kepler)
- [Kiali](#kiali)
- [Grafana](#grafana)
- [PostgresSQL](#postgres)
- [Airflow](#airflow)

Once changed the kustomization file, remember to push the modification to the repo to inform ArgoCD about the new configuration.

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



### PostgresSQL <a name="postgres"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/kustomization.yaml) used by argocd has the line ``- postgres`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- postgres
...
```

### Airflow <a name="airflow"/>

Be sure that, **on your repo** the [kustomization file](../deployment/mesh-infra/argocd/projects/kustomization.yaml) used by argocd has the line ``- airflow`` uncommented. E.g.:

```bash
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- project.yaml
...
- airflow
...
```