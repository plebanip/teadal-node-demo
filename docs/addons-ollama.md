# Ollama 

[Ollama](https://ollama.com/) is a tool that allows running the Large Language Models (LLMs) on regular servers, with or without the GPU enablement. In TEADAL, Ollama is used to power the [Automatic SFDP Generator (ASG)](./tools-asg.md).

## Dependencies

No dependencies are known at the moment but this can change, stay tuned :-).

## Preliminary steps

For the basic operations of the Automatic SFDP Generator (ASG), no special configuration is required.
This can change if, for example, it will be decided to:
- run `ollama` in its own namespace 
- allow users to influence what model (LLM) the service will bootstrap with (currently, the ASG uses the `granite-code:20b`)
- have additional TEADAL tools and services, beyond the ASG, to interract with `ollama` addon.

## Enabling the addon on a TEADAL Node

1. Assuming the node is created as described in [InstallTeadalNode](./InstallTeadalNode.md) and is backed by an up-to-date fork of the [Teadal.Node](https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node) repository, uncomment the `ollama` lines in the following files:

    1. In the [`kustomize` file for the `argocd` project for the `plat-app-services`](../deployment/mesh-infra/argocd/projects/plat-app-services/kustomization.yaml)
    1. In the [`kustomize` file for the local PVCs](../deployment/mesh-infra/storage/pvc/local/kustomization.yaml)
    1. In the [`kustomize` file for the `plat-app-services`](../deployment/plat-app-services/kustomization.yaml)

1. Commit the above changes to the repo/branch that ArgoCD is synchronised with:
    ```sh
    git add .
    git commit -m "enable ollama on this node"
    git push <remote-node-is-synced-with> <branch-node-is-synced-with>
    ```

1. Wait for ArgoCD to pick-up the changes or force the changes using ArgoCD UI or CLI.


## Checking the addon

1. Check that all the required objects are running:

    ```sh
    kubectl get pvc # should have 'ollama-pvc' in 'Bound' status
    kubectl get svc # should have 'ollama' with ClusterIP and 11434/TCP port
    kubectl get pod # should have `ollama-<xxxx> in 'Running' status 
    ```
    If this not the case, there is a need to check deeper. You can either investigate yourself or turn to TEADAL tech team for help.

2. Check `ollama` is serving and accessible from the node:

    ```sh
    # service ready
    $ curl localhost/ollama 
    Ollama is running   

    # correct model is loaded
    $ curl localhost/ollama/api/tags | jq
    {
        "models": [
            {
                "name":"granite-code:20b",
                "model":"granite-code:20b",
                ...
            }
        ]
    }
    ```

    You can also try to invoke other `ollama` APIs, e.g., to pull more models and/or answer questions, etc., both with `curl` or programmatically.

1. Check `ollama` is accessible from outide the node (to be called from other nodes in the federations of from the TEADAL developer's workstations if needed). For this you can repeat the same `curl` commands as above but replace the `localhost` with the `IP` address of the TEADAL node where the `ollama` service is enabled.

