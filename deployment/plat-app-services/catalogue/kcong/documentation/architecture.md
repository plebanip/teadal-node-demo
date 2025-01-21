# System Architecture Documentation

## Overview

This document provides a comprehensive overview of the Knowledge Catalogue system architecture deployed within our TEADAL MicroK8s cluster. The deployment leverages multiple nodes, routing layers, and service configurations to ensure scalability, security, and ease of access. Key architectural components include edge nodes, a central node (Teadal), Istio for service routing, and Nginx for internal routing. Additional services such as Keycloak, MinIO, and Django are included to support the backend, authentication, and storage needs.

---

## Table of Contents

1. [System Architecture Overview](#system-architecture-overview)
2. [Network Configuration](#network-configuration)
3. [MicroK8s Cluster Details](#microk8s-cluster-details)
4. [Routing Architecture](#routing-architecture)
    - [Istio Configuration](#istio-configuration)
    - [Nginx Configuration](#nginx-configuration)
5. [Service Details](#service-details)
    - [Backend Services](#backend-services)
    - [Frontend Services](#frontend-services)
    - [Authentication Services](#authentication-services)
    - [Storage Services](#storage-services)
6. [Environment Variables](#environment-variables)
7. [Diagrams](#diagrams)

---

## System Architecture Overview

### Components:
- **Edge Node**:
  - Publicly accessible via the internet.
  - IP Address: `131.175.120.210`
  - Acts as an entry point for external requests.

- **Teadal Node**:
  - Internal node within the Polimi network.
  - IP Address: `10.75.4.58`
  - Hosts the MicroK8s cluster.

- **MicroK8s Cluster**:
  - Hosts all application services, including Istio for routing and internal Nginx for service management.
  - Services include Keycloak, Django backend, MinIO, and frontend files.

---

## Network Configuration

### Edge Node:

- Public-facing entry point.
- Configured domains:
  - `baseline.teadal.ubiwhere.com`
  - `minio.baseline.teadal.ubiwhere.com`
  - `industry.teadal.ubiwhere.com`
  - `mobility.teadal.ubiwhere.com`
- Routes most traffic (`/`) to Istio, which determines the appropriate internal routing based on VirtualService configurations.


### Teadal Node:
- Internal IP: `10.75.4.58`
- Connects external requests from the edge node to the MicroK8s cluster (via ports that are mapped from inside the cluster to the ports of the node via NodePort)

---

## MicroK8s Cluster Details

### Namespaces:
- **Catalogue Namespace**:
  - Contains our services: Django, Keycloak, MinIO, and the Nginx internal router.

### Pods:
The catalogue namespace includes the following pods:
- Blazegraph (RDF Graph)
- BPMN (Camunda)
- Redis
- Django (backend services)
- Keycloak (authentication)
- MinIO (object storage)
- PostgreSQL (database)
- Nginx (internal routing)

---

## Routing Architecture

### Istio Configuration

#### Gateways:
- **Teadal Gateway**: Routes external traffic to the appropriate MicroK8s services.

#### Virtual Services:
Handles routing for various services, including:
- `/argocd`: Redirects to the ArgoCD UI.
- `/catalogue`: Handles frontend and backend services.
- `/...`: There are other mappings but for our use case we focus on /catalogue

#### URL Rewriting:
- `/catalogue` is rewritten to `/` when forwarded to internal Nginx.

### Nginx Configuration

#### Edge Node Nginx:
- Proxies requests from the public IP (`131.175.120.210`) to the internal Teadal node.

#### Internal Nginx:
- Runs within the `catalogue` namespace in the MicroK8s cluster.
- Routes traffic to:
  - Keycloak: `/keycloak/`
  - Django API: `/api/`
  - Admin: `/f**************/`
  - Static files: `/static/`
  - Media files: `/media/`

---

## Service Details

### Backend Services
- **Django**:
  - Hosts the API and admin interfaces.
  - Internal URL: `http://django.catalogue.svc.cluster.local:8000`

- **Redis**:
  - Acts as a caching and messaging service.

- **Blazegraph**:
  - Provides RDF data storage and query capabilities.

### Frontend Services
- **Nginx**:
  - Serves static files and frontend assets.

### Authentication Services
- **Keycloak**:
  - Provides user authentication and management.
  - Internal URL: `http://keycloak.catalogue.svc.cluster.local:8080/catalogue/keycloak/`

### Storage Services
- **MinIO**:
  - Handles object storage with signed URLs for direct file access.
  - Public domain: `http://minio.baseline.teadal.ubiwhere.com`

---

## Environment Variables

Some of the environment variables we set

### General Settings:
- `DOMAIN_NAME=baseline.teadal.ubiwhere.com`
- `EXTERNAL_FRONTEND_URL=http://baseline.teadal.ubiwhere.com/catalogue/`


### Keycloak Configuration:
- `KEYCLOAK_BASE_URL=http://keycloak.catalogue.svc.cluster.local:8080/catalogue/keycloak/`

### MinIO Configuration:
- `MINIO_DOMAIN=minio.baseline.teadal.ubiwhere.com`
- `MINIO_ROOT_USER=admin`
- `MINIO_ROOT_PASSWORD=XzUi7sLg*$og7A*s!bA^u2Dg`

---

## Diagrams

TODO

---

## Conclusion

This documentation outlines the system architecture for the Knowledge Catalogue. Detailed configurations and diagrams ensure a clear understanding of the network, services, and deployment setup. Use the provided environment variables and configurations to manage and troubleshoot the system effectively.

