#!/bin/bash

REGISTRY="${1:-ghcr.io/converged-computing}"
HERE=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT=$(dirname ${HERE})

# Go to the script directory
cd ${ROOT}

# These build each of the images. The sidecar is separate from the other two in src/
make build-all REGISTRY=${REGISTRY} # SCHEDULER_IMAGE=fluxqueue # SIDECAR_IMAGE=fluxqueue-sidecar

# This is what it might look like to push
# docker push ghcr.io/vsoch/fluxnetes-sidecar && docker push ghcr.io/vsoch/fluxnetes:latest

# We load into kind so we don't need to push/pull and use up internet data ;)
# kind load docker-image ${REGISTRY}/fluxnetes-sidecar:latest
kind load docker-image ${REGISTRY}/fluxqueue:latest
kind load docker-image ${REGISTRY}/fluxqueue-postgres:latest
kind load docker-image ${REGISTRY}/fluxqueue-scheduler:latest

# And then install using the charts. The pull policy ensures we use the loaded ones
#  --set scheduler.image=${REGISTRY}/fluxnetes:latest \
#  --set sidecar.image=${REGISTRY}/fluxnetes-sidecar:latest \
helm uninstall fluxqueue --namespace fluxqueue-system || true
helm install \
  --set postgres.image=${REGISTRY}/fluxqueue-postgres:latest \
  --namespace fluxqueue-system \
  --create-namespace \
  --set postgres.pullPolicy=Never \
        fluxqueue chart/
