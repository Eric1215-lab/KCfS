# SPDK CSI Driver

The SPDK CSI Driver exposes [SPDK](https://spdk.io/) logical volumes to
[Kubernetes](https://kubernetes.io/) through the
[Container Storage Interface (CSI)](https://github.com/container-storage-interface/spec).
It supports dynamically provisioned volumes, static volumes, and volume
snapshots through SPDK JSON-RPC.

## Features

- Kubernetes CSI driver (`csi.spdk.io`) written in Go.
- NVMe over Fabrics targets over TCP or RDMA.
- iSCSI targets.
- Dynamic provisioning backed by SPDK logical volumes.
- Static PersistentVolume and PersistentVolumeClaim workflows.
- Volume snapshots.
- Deployment with raw Kubernetes manifests or the bundled Helm chart.

## Prerequisites

- A Kubernetes cluster compatible with the Kubernetes dependencies declared in
  `go.mod` (currently v1.25).
- An SPDK target with a logical-volume store and an accessible JSON-RPC HTTP
  proxy.
- `kubectl` configured for the target cluster.
- Docker, Go 1.19+, and `make` when building the driver image locally.
- NVMe-oF initiator support on Kubernetes nodes when using NVMe-TCP or RDMA.

For a local test target, see [deploy/spdk/README.md](deploy/spdk/README.md).

## Quick start with Helm

Build the driver image first. The default chart image is
`spdkcsi/spdkcsi:canary` with `imagePullPolicy: Never`, so the image must be
available to the Kubernetes nodes.

```bash
make image

helm upgrade --install spdk-csi ./charts/spdk-csi \
  --namespace spdk-csi --create-namespace
```

Before deploying to a remote SPDK target, update the following values in
[`charts/spdk-csi/values.yaml`](charts/spdk-csi/values.yaml):

- `csiConfig.nodes`: SPDK target name, JSON-RPC URL, transport type, and target
  address.
- `csiSecret.rpcTokens`: JSON-RPC credentials; each `name` must match the
  related `csiConfig.nodes` entry.
- `image.spdkcsi`: image repository, tag, and pull policy for your registry.

To remove the Helm release:

```bash
helm uninstall spdk-csi --namespace spdk-csi
```

## Deploy with Kubernetes manifests

The manifest deployment uses the files under `deploy/kubernetes`.
Configure the SPDK endpoint and credentials in `config-map.yaml` and
`secret.yaml`, then apply the deployment:

```bash
cd deploy/kubernetes
./deploy.sh

kubectl get pods
kubectl apply -f testpod.yaml
```

Remove the test workload and driver resources when finished:

```bash
kubectl delete -f testpod.yaml
./deploy.sh teardown
```

For a multi-node example, including SPDK target setup and configuration, see
[docs/multi-node.md](docs/multi-node.md).

## Static volumes

Existing SPDK logical volumes can be exposed to Kubernetes as static PVs.
Examples for NVMe-oF and iSCSI are in
[docs/static-pvc.md](docs/static-pvc.md). Static PVs must use the `Retain`
reclaim policy because the driver does not delete the underlying SPDK logical
volume.

## Development

```bash
# Build a Linux driver binary in _out/spdkcsi
make spdkcsi

# Run Go module verification and unit tests
make test

# Run the full local build, lint, and test targets
make
```

Additional targets include `make e2e-test` for end-to-end tests and `make
helm-test` for a Helm installation check. End-to-end tests require a prepared
Kubernetes and SPDK environment.

## Repository layout

| Path | Purpose |
| --- | --- |
| `cmd/` | Driver entry point. |
| `pkg/` | CSI driver, SPDK integration, and shared utilities. |
| `charts/spdk-csi/` | Helm chart. |
| `deploy/kubernetes/` | Raw Kubernetes manifests and deployment script. |
| `deploy/spdk/` | Local SPDK target container setup. |
| `docs/` | Deployment and static-volume guides. |
| `e2e/` | End-to-end test suite and sample manifests. |
