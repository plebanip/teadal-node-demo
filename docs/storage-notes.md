# Storage Provisioning on the TEADAL.Node

When deploying a TEADAL.node you will be faced with the choice of how to provision storage. 

In the node documentation, you may find mentions of three methods to accomplish this, which this document will now try to summarize. It should be noted that these are by no means the only ways to handle this issue. If you are an experienced cluster manager, you may have preference for other, in which case, just go for what you like.

Before diving into the methods, lets do a quick summary of the current storage needs of the TEADAL.Node.

Running `kubectl get pvc -A` on a running Teadal.Node, will give you an output much like this one:

```bash
airflow          data-airflow-postgresql-0   Bound    tv-teadal-8   10Gi       RWO            local-storage   88d
airflow          logs-airflow-triggerer-0    Bound    tv-teadal-6   10Gi       RWO            local-storage   88d
airflow          logs-airflow-worker-0       Bound    tv-teadal-4   10Gi       RWO            local-storage   88d
airflow          redis-db-airflow-redis-0    Bound    tv-teadal-2   10Gi       RWO            local-storage   88d
default          keycloak-pvc                Bound    tv-teadal-7   10Gi       RWO            local-storage   88d
default          postgres-pvc                Bound    tv-teadal-1   10Gi       RWO            local-storage   88d
kubeflow         mysql-pv-claim              Bound    tv-teadal-9   20Gi       RWO            local-storage   42d
minio-operator   0-teadal-teadal-0           Bound    tv-teadal-3   10Gi       RWO            local-storage   88d
```

We can observe that, currently, the baseline has 8 storage claims, by default. They are:

* Four 10GB claims for airflow
* One 10GB claim for keycloak
* One 10GB claim for postgres
* One 10GB claim for minio
* One 20GB claim for kubeflow

With this in mind, we can now go over on how to actually provide the needed storage for these services.

## The Microk8s way

Microk8s has an addon for storage provisioning: `hostpath-storage`. 

It can be enabled by running `microk8s enable hostpath-storage`.

This addon completely streamlines the process of provisioning, making it a very useful tool. While it is not really scalable, it works great in single-node deployments, which is our case for the first iteration.

If `hostpath-storage` is active, there is no need to manually create the PV, the addon will do so when the PVC is deployed. 

When using this solution, you should set the StorageClass of every PVC as `microk8s-hostpath`, by default.
If you want to use custom StorageClasses, that should be possible as well, just make sure you set the provisioner field as: 

```yaml
provisioner: microk8s.io/hostpath
```

## The Local Storage Way

Another way to provision storage is DIY-ing local storage. This a valid method since we only have one node. The process to accomplish do this is described on the documentation [here](https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node/-/blob/main/docs/bootstrap/mesh.md#k8s-storage).

This method has a drawback in comparison to the microk8s way because you have to create the partitions/directories used manually, as well as the PVs, which microk8s `hostpath-storage` automates for you. 

When using this method, you should set StorageClass in your PVC's as `local-storage`, by default.

## The DirectPV Way

DirectPV is a k8s storage provisioner, that can be used in production environments. Since we are using one node only, it should not be necessary to use this solutions, but it's certainly an option. 

DirectPV installation is documented [here](https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node/-/blob/main/deployment/mesh-infra/storage/directpv/kustomization.yaml?ref_type=heads).

If using this solution, the StorageClass of each PVC has to be set to `directpv-min-io`, by default.
