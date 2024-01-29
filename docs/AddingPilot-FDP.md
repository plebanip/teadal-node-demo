## How to add a new Federated Data Product (FDP) to a TEADAL node

According to the TEADAL approach, a FDP is composed of two elements:
* a REST based service exposing data available in the node
* a set of rules expressed in OPA rego to control the access to the methods exposed by the service and the content of the data shared

### Concerning the REST based service

The REST based service must be located in a container available on a registry. To know how to build a container there are many guides on the web:
* the official source https://docs.docker.com/engine/reference/commandline/image_build/
* a short tutorial https://docs.docker.com/get-started/02_our_app/ 
* a nice guide for applications developed with nodejs https://code.visualstudio.com/docs/containers/quickstart-node (please note that no restrictions are given to the platform used to deploy and run your code as long as they expose a REST API)

The image of the container should be available on a registry like [DockerHub](https://hub.docker.com). 
If you are a member of the TEADAL team, the project makes available two registries related to two projects in our gitlab repo:
* a public one https://gitlab.teadal.ubiwhere.com/teadal-public-images. Here you are required to be authenticated to push your image, but pulling images is free
* a private one https://gitlab.teadal.ubiwhere.com/teadal-images. Here you are required to be authenticated for both pushing and pulling images

Once the image has been published, it is required to create (i) two files to inform ArgoCD about the new FDP to be deployed, and other (ii) two files to inform microk8s about how to deploy.

#### Inform ArgoCD

Under `deployment/mesh-infra/argocd/projects/pilot-services` create a new folder for your FDP. Select a name associated to your project. Hereafter, we refer to this name as <NAME_OF_YOUR_FDP>.

```bash 
mkdir deployment/mesh-infra/argocd/projects/pilot-services/<NAME_OF_YOUR_FDP>
```
 Then copy there the two template files [app.yaml](./templates/argocd/app.yaml) and [kustomization.yaml](./templates/argocd/kustomization.yaml). The latter does not require any changes. About the former, simply change the name of the placeholder with the name selected for your FDP

 ```yaml
 - op: replace
  path: /metadata/name
  value: <NAME_OF_YOUR_FDP>
- op: replace
  path: /spec/source/path
  value: deployment/pilot-services/<NAME_OF_YOUR_FDP>
- op: replace
  path: /spec/project
  value: pilot-services
```

In this way, when the repository is updated, ArgoCD knows that there is a new FDP to be deployed in your node.

#### Inform Microk8s

Under `deployment/pilot-services` create a new folder for your FDP named as <NAME_OF_YOUR_FDP>.

```bash 
mkdir pilot-services/<NAME_OF_YOUR_FDP>
```

Then copy there the two template files [app.yaml](./templates/microk8s/base.yaml) and [kustomization.yaml](./templates/microk8s/kustomization.yaml). Also in this case, only the latter requires some changes. See them step by step.

At the beginning the manifest define the k8s service. Please note that regardless of the port that is used to expose the service, because of Istio, all the services will be available on port 80.

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app: <NAME_OF_YOUR_FDP>
  name: <NAME_OF_YOUR_FDP>
spec:
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  selector:
    app: <NAME_OF_YOUR_FDP>
```

In the second section, there is the deployment instructions for k8s. Here it is required to substitute the placeholder <NAME_OF_YOUR_FDP> accordingly. Here it is also needed to indicate the <NAME_OF_YOUR_IMAGE> which value depends on where the image has been pushed. Information about the container could be completed with specific parameters. For instance, in this case, there are the credential to access to the minio instance in the node to allow the FDP to be connected with.

``` yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: <NAME_OF_YOUR_FDP>
  name: <NAME_OF_YOUR_FDP>
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <NAME_OF_YOUR_FDP>
  template:
    metadata:
      labels:
        app: <NAME_OF_YOUR_FDP>
    spec:
      containers:
        - name: <NAME_OF_YOUR_FDP>
          image: <NAME_OF_YOUR_IMAGE>
          ports:
          - containerPort: 8080
          env: #  Not ideal to hardcode the minio credentials, but we can't mount secrets from a different namespace so ...
          - name: MINIO_HOST
            value: "teadal-teadal-0.teadal-hl.minio-operator.svc.cluster.local"
          - name: MINIO_PORT
            value: "9000"
          - name: MINIO_ACCESS_KEY
            value: LTXG5CVY0MXLE0WO75Q6
          - name: MINIO_SECRET_KEY
            value: IKxps9GTiPGBZaK6BRmeF434lGZCNx0XG3sa4PLE          
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
```

Finally, the manifest has to define the VirtualService used by Istio to create the proxy. Yet, here it is enough to substitute the placeholder <NAME_OF_YOUR_FDP>.

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: "<NAME_OF_YOUR_FDP>-virtual-service"
spec:
  gateways:
  - "teadal-gateway"
  hosts:
  - "*"
  http:
  - match:
    - uri:
        prefix: /<NAME_OF_YOUR_FDP>/
    - uri:
        prefix: /<NAME_OF_YOUR_FDP>
    rewrite:
      uri: /
    route:
    - destination:
        host: <NAME_OF_YOUR_FDP>.default.svc.cluster.local
        port:
          number: 8080

```


### Concerning the OPA Rego file

The FDP has to be associated with two OPA Rego files which express the access control rules checked everytime a request for data is addressed to your FDP. For more information about OPA Rego please refer to the [official guide](https://www.openpolicyagent.org/docs/latest/policy-language/).

A first rego file is the entry point for the TEADAL node to be aware of your rules and should be organized like this [service rego template](./templates/service.rego)

```
#
# Example policy for the sync dummy FDP.
#

package <NAME_OF_YOUR_FDP_FOR_REGO>.service

import input.attributes.request.http as http_request
import data.authnz.envopa as envopa
import data.config.oidc as oidc_config
import data.<NAME_OF_YOUR_FDP_FOR_REGO>.rbacdb as rbac_db


default allow := false

allow = true {
    user := envopa.allow(rbac_db, oidc_config)

    # Put below this line any service-specific checks on e.g. http_request

}
```

You are required at least to change the placeholder <NAME_OF_YOUR_FDP_FOR_REGO>. The value could be the same as <NAME_OF_YOUR_FDP> defined in advance, but in case it contains special characters like `_` `?` `-` these can create confusion to OPA, so it is better to remove them. At this stage no other modifications are required. In next steps, it will be possible to specify additional controls.


A second rego file contains the actual rules. A suggested structure is reported in this [template](./templates/rbacdb.rego). Here there is the access control logic. Let examine the structure step by step.

First of all rename the package with the <NAME_OF_YOUR_FDP_FOR_REGO> you have specified in the other rego file.

```#
package <NAME_OF_YOUR_FDP_FOR_REGO>.rbacdb

import data.authnz.http as http
```

Defines the role of your FDP. In this case, we have two basic roles

```
# Role defs.
product_owner := "product_owner"
product_consumer := "product_consumer"
```

Map each role to a list of permission objects. Each permission object specifies a set of allowed HTTP methods for the Web resources identified by the URLs matching the given regex. To Web resources must refer to the FDP path. Thus, the <NAME_OF_YOUR_FDP> must be replaced with the value in the REST service manifest. In this case, the value must be the same used in definition of the REST service since it is related to the PATH created to access to your service. Then, you can add any rule you want as long as they are expressed with the right syntax.

```
role_to_perms := {
    product_owner: [
        {
            "methods": http.do_anything,
            "url_regex": "^/<NAME_OF_YOUR_FDP>/.*"
        }
    ],
    product_consumer: [
        {
            "methods": http.read,
            "url_regex": "^/<NAME_OF_YOUR_FDP>/<SUBPATH>"
        }
    ]
}
```

Finally, Map each user to their roles. User must exists in the keycloak. Thus it is important to register them in advance. See [AddingUsers](./AddingUsers.md).

```
user_to_roles := {
    "jeejee@teadal.eu": [ product_owner, product_consumer ],
    "sebs@teadal.eu": [ product_consumer ]
}
```

Please note that, based on the complexity of the access control of your FDP, you can create more that one OPA rego file. What it is important is that at least one of it is named 'rbacdb.rego' to create a link with the rest of the system.


### Deploying the FDP

After setting up everything you need to update the repo with your modification

```bash
git push
```

After some minutes, ArgoCD fecthes the changes in the repo and deploy the FDP accordingly. If everything goes well, new pods will appear running the command

```bash
kubectl get pod -A
```

If everything goes very well, your service will be available at the url ```localhost/<NAME_OF_YOUR_FDP>```