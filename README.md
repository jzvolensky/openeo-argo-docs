# Eurac openeo argoworkflows docs

This repository contains documentation for deploying openeo-argoworkflows by EODC at Eurac.

## Table of Contents

TODO later

## Installation

There are several automated helpers available

**Please note that only x86_64 Linux* systems are currently supported.**

*Tested on `Ubuntu 22.04 LTS`*

*Recommended to only use internally on the Eurac infrastructure. Unexpected behavior may occur on other systems.*

### install-tools

The `install-tools` automates the installation of common tools required for working with OpenEO and Argo Workflows.

It checks for the presence of each tool and installs it if not found.

The tools include:

- kubectl
- helm
- minikube
- argo cli

to build it yourself you need `go 1.25`. Somewhat older versions should work too but they are not tested.

Visit [golang website](https://golang.google.cn/learn/)

Build it with:

```sh
go build -o install-tools ./install-tools
```

Run the installer:

```sh
./install-tools
```

If everything goes well you should see something like:

```sh
Summary (versions):
✔ kubectl: Client Version: v1.29.2
✔ helm: v3.14.1+ge8858f8
✔ minikube: minikube version: v1.32.0
✔ argo: argo: v3.7.1

🎉 All tools are ready to use!
```

### Helm setup

**Go into the charts/eodc/openeo-argo directory first!**

Add the following helm repositories:

```sh
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo add argo https://argoproj.github.io/argo-helm
helm repo add dask https://helm.dask.org
helm repo update
helm dependency build
```

### Start Minikube

```sh
minikube start
```

Create a namespace for the OpenEO deployment:

```sh
kubectl create ns openeo
```

### Deploy OpenEO Argo Workflows

Clone the repository containing the helm chart:

```sh
git clone git@github.com:eodcgmbh/charts.git
```

The default `values.yaml` comes preconfigured for a local test deployment.

There is no need to change anything for a first test deployment.

```sh
helm install openeo -n openeo -f values.yaml .
```

Wait and check the deployment status:

```sh
kubectl get pods -n openeo
```

First time setup took quite some time to pull all the images.

To access the OpenEO API you need to setup port forwarding:

```sh
kubectl port-forward -n openeo svc/openeo 8000:80

# And access the API at:
http://127.0.0.1:8000/openeo/1.1.0/
```
