[![Actions Status](https://github.com/symcn/mesh-operator/workflows/go-build/badge.svg)](https://github.com/symcn/mesh-operator/actions?query=workflow%3Ago-build)
[![Actions Status](https://github.com/symcn/mesh-operator/workflows/pre-commit/badge.svg)](https://github.com/symcn/mesh-operator/actions?query=workflow%3Apre-commit)
# mesh-operator
It is in charge of all things about implementation of Service Mesh.

What are our tasks?

1. Retrieving the data of service & instances from a registry center such as Nacos, zookeeper. 
2. Generating the istio's CRD through various data such as registry data & configurations.
3. Make MOSN can support registry request of a dubbo provider.

## Installing and running

### Build

Using Docker unified compilation(with version information):

```shell
$ make docker-build
```

Direct compilation:

```shell
$ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o build/_output/bin/mesh-operator -ldflags "-s -w" cmd/mesh-operator/main.go
```

Build docker image and push:

```shell
$ make docker-push
```

You can customize the docker repository by modifying the `IMG_ADDR` variable in the [Makefile](./Makefile).

### Deploy

Install CRD:

```shell
$ cd deploy/crds
$ kubectl apply -f mesh.symcn.com_appmeshconfigs_crd.yaml -f mesh.symcn.com_istioconfigs_crd.yaml -f mesh.symcn.com_meshconfigs_crd.yaml -f mesh.symcn.com_servicemeshentries_crd.yaml
```

Install CRD no validations:
```shell
$ cd deploy/crds
$ kubectl apply -f simple_appmeshconfigs_crd.yaml -f simple_istioconfigs_crd.yaml -f simple_meshconfigs_crd.yaml -f simple_servicemeshentry_crd.yaml
```

Create ClusterRole, ClusterRoleBinding and ServiceAccount:
```shell
$ cd deploy
$ kubectl apply -f role.yaml -f role_binding.yaml -f service_account.yaml
```

Deploy adapter:
```shell
$ kubectl apply -f deploy/adapter.yaml
```

Deploy controller:
```shell
$ kubectl apply -f deploy/controller.yaml
```


## Getting support

- Some docs comming soon

## Features

### Supported

### Planned
