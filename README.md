# nginx-operator
The nginx-operator enables Kubernetes users to quickly deploy nginx instances in a Kubernetes cluster. The operator handles the management of the instances and the way they are reached through the public endpoint of the cluster.

## Description

The operator has three configuration options, all mandatory:

- `replicas`: this is used to configure the ReplicaSet of the Deployment used to manage the Nginx pods
- `image`: this is the name that is used to define the Nginx image that will be deployed
- `hostname`: this name is the publicly routable DNS name that can be used to reach the cluster environment, used to configure the Ingress

This operator handles the management of three components:
- Deployment:
The nginx instances are operated through a Deployment, where the number of instances in the ReplicaSet of the Deployment is the `replicas` value of the Nginx manifest.
- Service:
The Pods are reachable through a Service, which statically binds port 80 of the pods' container to port 80 of the service. The Service's IP is `ClusterIP`, not reachable outside the cluster.
- Ingress:
The way traffic towards the nginx instances is routed is through an Ingress, which routes the root '/' path through HTTPS using TLS to port 80 of the Service defined above.

## Prerequisites

For this infrastructure to work, there are three prerequisites:

1. The infrastructure has a publicly routable IP address and a corresponding DNS entry. This is required so that the `hostname` in the Nginx manifest is a valid internet address that can be verified during the certificate requisition. This guide assumes the use of Let's Encrypt and its staging API.

2. The following resources are also present in the cluster:

- NGINX Ingress Controller (the operator assumes that the ingress operator is using nginx, otherwise its preset configuration won't work)
- cert-manager (the operator does not make assumptions about how the TLS certificates are managed, but it has requirements based on naming)

For installation, the following will provide the minimal configuration that allows the operator to function:

- Install the NGINX Ingress Controller by following the [environment-specific guide here](https://github.com/kubernetes/ingress-nginx/blob/main/docs/deploy/index.md) or through this command:

```sh
# NGINX Ingress Controller
# For Helm users
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace
# For kubectl users with manifests (mind the versioning)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.6.4/deploy/static/provider/cloud/deploy.yaml
```

- You will also need the components of cert-manager to be in place through following the [steps here](https://cert-manager.io/docs/installation/), or by running this command:

```sh
# cert-manager (mind the versioning)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

3. The Issuer or ClusterIssuer resource is deployed to the cluster (apply your use case):

```yaml
# Resource type depends on choice, the difference is
# that Issuer is restricted to the namespace it is deployed in.

apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging # Name should stay, the controller depends on it
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: <your email address>
    privateKeySecretRef:
      name: letsencrypt-staging # Name should stay, the controller depends on it
    solvers:
      - http01:
          ingress:
            class: nginx # Nginx should stay, as per the controller prerequisite

# OR

apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging # Name should stay, the controller depends on it
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: <your email address>
    privateKeySecretRef:
      name: letsencrypt-staging # Name should stay, the controller depends on it
    solvers:
      - http01:
          ingress:
            class: nginx # Nginx should stay, as per the controller prerequisite
```

The cert-manager will watch the Ingress resource that has `letsencrypt-staging` set as its certificate issuer, and when a TLS secret is defined, such as `letsencrypt-staging`, it will initiate the requisition of the certificate through the connected (Cluster)Issuer. That resource will handle the CertificateRequest and Certificate resources, and automatically manage renewals. As the secret referenced in the TLS section is what's used for naming the Certificate, the connection between the Ingress and the Certificate is dynamic.

## Quick Start

1. Create your Kubernetes cluster and set up a publicly accessible DNS name.

2. Use the default namespace for all operations that don't require their own namespace.

3. Install the prerequisites: NGINX Ingress Controller and cert-manager.

4. Create the (Cluster)Issuer resource.

5. Create your connection to your cluster, ensure the presence of a configuration file at '~/.kube/config'.

6. Clone this repository. Using the Makefile, deploy the operator to your cluster.

```sh
make deploy IMG=ghcr.io/rubrumvulpes/nginx-operator:main
```

7. Create the manifest for the Nginx resource and apply to your cluster.

```yaml
apiVersion: webserver.cisco.davidkertesz.hu/v1
kind: Nginx
metadata:
  labels:
    app.kubernetes.io/name: nginx
    app.kubernetes.io/instance: nginx-sample
    app.kubernetes.io/part-of: nginx-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: nginx-operator
  name: nginx-sample
spec:
  image: "nginx:latest"
  replicas: 3
  host: "<your routable domain>"
```
