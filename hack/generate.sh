#!/usr/bin/env bash

set -euo pipefail

echo "Running controller-gen inside Docker..."

docker run --rm -v "$(pwd)":/workspace -w /workspace golang:1.23-alpine sh -c '
  mkdir -p ./deploy/fake-network-operator/crds && \
  go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.2 && \
  controller-gen crd paths=./api/... output:crd:dir=./deploy/fake-network-operator/crds && \
  controller-gen object paths=./api/...
'

echo "Code generation complete."
