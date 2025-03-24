# Configure Keycloak to include role information in JWT

In case you need to specify access controls based on the roles and not only on the username, Keycloak must be configured to include the roles information in the released JWTs.

If you do not have already defined your roles, please have a look to the [official guide](https://www.keycloak.org/docs/latest/server_admin/index.html#assigning-permissions-using-roles-and-groups). Notably, the paragraphs "Creating a realm role" and "Assigning role mappings".


Once the roles have been defined, follow these steps:

1) Open Keycloak admin page (e.g., http://<node_IP>/keycloak)

2) Select the "teadal" realm

<img src="../images/select-realm.png" alt="Select realm" width="50%"/>

3) Select "Clients" and from the list appearing on the right "admin-cli"

<img src="../images/select-admincli.png" alt="Select client" width="70%"/>

4) Select "Client scopes" panel and from the list "admin-cli-dedicated"

<img src="../images/select-adminclidedicated.png" alt="Select client scope" width="70%"/>

5) Select "Scope" panel 

<img src="../images/select-scopepanel.png" alt="Select client scope" width="70%"/>

6) In this panel the set of roles you want to be included in the JWT must be added. If all of them already appear, then no need to do anything. Otherwise, proceed with the next steps

7) To add the roles, click on "Assign roles"

<img src="../images/select-assignrole.png" alt="Assign roles" width="50%"/>

8) Select all the roles you want to expose and click on "Assign"

<img src="../images/select-roles.png" alt="Select roles" width="70%"/>

9) The updated list of roles should appear

<img src="../images/listroles.png" alt="List roles" width="70%"/>

