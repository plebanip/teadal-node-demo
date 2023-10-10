# teadal.proto
Messing around with cloud infra for https://www.teadal.eu.

This is still very much a work in progress, but the code for a
(possible) cloud baseline is there. [Read this][docs] before diving
into the code. This is what the cloud prototype looks like at the
moment.

## Usage
To use this code in its current state to develope an FDP, fork this repo and than look at:
 - [/deployment/pilot-services/fdp-dummy/][fdp-dummy]
 - [deployment/mesh-infra/security/opa/rego/fdpdummy/][fdp-opa]
 - change [deployment/mesh-infra/argocd/projects/base/app.yaml](./deployment/mesh-infra/argocd/projects/base/app.yaml)

![Prototype tech stack.][dia.tech-stack]


[docs]: ./docs/README.md
[dia.tech-stack]: ./docs/tech-stack.svg
[fdp-dummy]: ./deployment/pilot-services/fdp-dummy/base.yaml
[fdp-opa]: ./deployment/mesh-infra/security/opa/rego/fdpdummy/service.rego