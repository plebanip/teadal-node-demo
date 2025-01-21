If you are experiencing with key cloak issues when you moved to the new version of tidal released at the end of dec 2024, here a short guide you have to look at AFTER the fork update has succeeded.

This guide is required as, in the new setting, keycloak no longer uses the internal DB but uses Postgres. For this reason, Postgres is not mandatory and it has been moved from 'plat-infra' to 'mesh-infra'.

1) Enter in the VM hosting the cluster corresponding to the node

2) Move to the repo folder clone

3) Update the repo `git pull`

4) Update the node.config command with the new version running `nix build`

5) Once the build has finished, remove the result folder link `rm -f result`

6) Enter in the shell `nix shell`

7) Check if you have the new version of the node.config running `node.config`. You are on the right path if the help returns in the COMMANDS sections all of these possibilities:

```
COMMANDS:
   basicnode-secrets  set/reset secrets for basic teadal node installation
   postgres-secrets   set/reset postgres secrets
   keycloak-secrets   set/reset keycloak secrets
   advocate           configure advocate tool
   pv                 generate pvs for your teadal node
   help, h            Shows a list of commands or help for one command
   ```

8) Now we need to create a Postgres user named 'keycloak' that will be used by keycloak to store the needed data. To do this we need to run `node.config --microk8s postgres-secrets`. This command will ask for two passwords: a) the password for the typical root users (Postgres) b) the password for the key cloak user

9) Now we need to inform Postgres about the new user. We need to run ``kubectl exec -it <postgres-pod-name> -- /bin/bash``. Once inside the container run ``psql --username postgres`` to enter in Postgres. Once in Postgres, execute, one by one, these four commands:

```
CREATE USER keycloak WITH PASSWORD 'the second password you specified in node.config'; 
  CREATE DATABASE keycloak;
  GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak; 
  ALTER DATABASE keycloak OWNER TO keycloak;
```

10) re-deploy keycloak running `kubectl delete deployment/keycloak` and `kustomize build deployment/mesh-infra/security/keycloak | kubectl -f -`

11) if everything worked, after a couple of minutes running the `Kubectl get pod -A` you can see the keycloak pod correctly running along with its sidecar (thus 2/2 must appear)