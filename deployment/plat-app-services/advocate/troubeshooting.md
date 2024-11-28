# Troubleshooting Advocate

## Reinstalling
In case you need to reinstall it delete advocate by deleting the applied kustomize file `kubectl delete -k .` to get rid of most services.
You also will have to clean up the Webhook using `kubectl delete MutatingWebhookConfiguration advocate-ambassador` and any residule files in IPFS.