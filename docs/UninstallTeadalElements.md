# Uninstall elements from a Teadal node 

## Uninstall Teadal tools, including FDPs, and SFDPs
The easiest way to remove the TEADAL tools, the FDPs, and SFDPs is to operate via argocd. This can be done using the web interface or using the ArgoCD CLI using the following commands.

In the VM shell run the command ``kubectl get endpoints -A`` to obtain the address of the argocd server.

```
...
argocd           argocd-server                             10.1.9.14:8080,10.1.9.14:8080                              59m
...
```

Then, login to the argocd server running the command ``argocd login <address>:8080`` (e.g., ``argocd login 10.1.9.14:8080``). The ArgoCD client will ask for the username (admin) and password of argocd that have been specified during the node installation.

Running the command ``argocd app list`` is possible to see the applications that can be managed. E.g.:

```
argocd/argocd          https://kubernetes.default.svc  default    mesh-infra         OutOfSync  Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/argocd                 integrate_thanos
argocd/dspn-webeditor  https://kubernetes.default.svc  default    plat-app-services  Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/plat-app-services/dspn-webeditor  integrate_thanos
argocd/httpbin         https://kubernetes.default.svc  default    plat-app-services  Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/plat-app-services/httpbin         integrate_thanos
argocd/keycloak        https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/security/keycloak      integrate_thanos
argocd/minio           https://kubernetes.default.svc  default    mesh-infra         OutOfSync  Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/storage/minio          integrate_thanos
argocd/opa             https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/security/opa           integrate_thanos
argocd/postgres        https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/storage/postgres       integrate_thanos
argocd/pvc             https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/storage/pvc            integrate_thanos
argocd/reloader        https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/security/reloader      integrate_thanos
argocd/routing         https://kubernetes.default.svc  default    mesh-infra         OutOfSync  Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/istio/routing          integrate_thanos
argocd/sc              https://kubernetes.default.svc  default    mesh-infra         Synced     Healthy  Auto        <none>      https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node.git  deployment/mesh-infra/storage/sc             integrate_thanos
```

In case you need to delete an app ``argocd app delete argocd/httpbin``

## Uninstall Istio add-ons

To uninstall the Istio add-ons the best way is to operate directly with kubectl.

### Kiali

``kubectl delete deployment kiali -n istio-system``

### Grafana

``kubectl delete deployment grafana -n istio-system``

### Jaeger

``kubectl delete deployment jaeger -n istio-system``

### Kepler

``kubectl delete deployment kepler -n istio-system``

### Prometheus 

If Prometheus have been installed WITH Thanos, then Thanos must be removed first

``kubectl delete deployment thanos-compact thanos-query  thanos-sidecar prometheus -n istio-system``

otherwise you need to remove only Prometheus

``kubectl delete deployment prometheus -n istio-system``

