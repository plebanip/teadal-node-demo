# Troubleshooting Advocate

#### Advocate deployment

In path */deployment/plat-infra-services/advocatel* , there are files related to deploy Advocate service. 

In order  to deploy Advocate, you need to update the kustomization.yaml in */deployment/plat-infra-services/advocatel* directory with actual values of your deployment environment :

 * Make sure to update this command with the IP address of your blockchain server:
 * The name of blockchain host should be "teadal.blockchain". Please do not alter the name.

```bash
patches:
- target:
    kind: Deployment
    name: advocate
  patch: |-
    - op: add
      path: /spec/template/spec/hostAliases
      value:
      - ip: "<TEADAL_BLOCKCHAIN_IP>"
        hostnames:
        - "teadal.blockchain"
```
There is also some addresses under the configMapGenerator property that can be replaced with custom values of your deployment:

```bash
configMapGenerator:
- name: advocate-config
  literals:
  - ADVOCATE_IPFS_CONNECTION=<IPFS_CONNECTION_ADDRESS>
  - ADVOCATE_JAEGER_QUERY_ADDRESS=tracing.istio-system.svc.cluster.local:16685
  - ADVOCATE_SELF_ADDRESSES=<ADVOCATE_ADDRESSES>,...
  - ADVOCATE_PUBLIC_ADDRESS=https://localhost/advocate

```

In case that you are using the embeded IPFS, you need to address ipfs.yaml file in resources in *Kustomization.yaml* and the  default value of *ADVOCATE_IPFS_CONNECTION* config which is *http://ipfs.trust-plane.svc.cluster.local:5001/api/v0/*  does not need to be changed. Otherwise, put address of your ipfs service in *<IPFS_CONNECTION_ADDRESS>* and comment the ipfs.yaml under resources property. *<ADVOCATE_ADDRESSES>*  is the internal address of Advocate service in deployment environment and *<ADVOCATE_PUBLIC_ADDRESS>* is public address to access the Advocate. In this file there is also tracing service address (JAEGER) which points to the tracing service in the istio-system name space on default port 16685.



Check the */deployment/plat-infra-services/kustomization.yaml* file to make sure that the advocate is in the list of resources and the run below command to deploy Advocate in a namespace called *trust-plane* .

```bash
kustomize build plat-infra-services/advocate/kustomization.yaml | kubectl apply -f -
```

Note, that for Advocate to work flawlessly, you need to have permissions to create cluster resources and namespaces.
Furthermore, you'll need to have access to an ETH network, either testing or mainnet, with a valid account and some ETH to pay for the transactions. Moreover, you either have to deploy or have access to IPFS node to store the files. You can uncomment the IPFS deployment in the `kustomization.yaml` file to deploy an IPFS node.
We assume your cluster is running a recent version of Jaeger that this node can reach. Before Advocate will work you will need to configure all needed secrets, variables releated to the Advocate blockchain such as wallet private key, VM key and Ethereum Remote Procedure Call (RPC) Address. For that follow the steps in the [K8s secrets](#k8s-secrets) section.

Check pods that are in Trust-plane namespace:

```bash
kubectl get pods -n trust-plane
```

![screenshot](/././docs/images/trust-plane-namespace-podes.png)

Do not forget to check Advocate pod log to make sure that it is up and running:

```bash
kubectl logs <advocate-pod-name> -n trust-plane
```

![screenshot](/././docs/images/advocate-pod-log.png)


#### Reinstalling
In case you need to reinstall it delete Advocate by deleting the applied kustomize file `kubectl delete -k .` to get rid of most services.
You also will have to clean up the Webhook using `kubectl delete MutatingWebhookConfiguration advocate-ambassador` and any residule files in IPFS.