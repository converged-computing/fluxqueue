#!/bin/bash

REGISTRY="${1:-ghcr.io/converged-computing}"
HERE=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT=$(dirname ${HERE})

# Go to the script directory, update helm charts
cd ${ROOT}
make helm

# These build each of the images. The sidecar is separate from the other two in src/
make build-all REGISTRY=${REGISTRY} # SCHEDULER_IMAGE=fluxqueue # SIDECAR_IMAGE=fluxqueue-sidecar

# We load into kind so we don't need to push/pull and use up internet data ;)
kind load docker-image ${REGISTRY}/fluxqueue:latest
kind load docker-image ${REGISTRY}/fluxqueue-postgres:latest
kind load docker-image ${REGISTRY}/fluxqueue-scheduler:latest

# And then install using the charts. The pull policy ensures we use the loaded ones
helm uninstall fluxqueue --namespace fluxqueue-system --wait || true

# So we don't try to interact with old webhook, etc.
sleep 5
helm install \
  --set controllerManager.manager.image.repository=${REGISTRY}/fluxqueue \
  --set controllerManager.manager.image.tag=latest \
  --set scheduler.image=${REGISTRY}/fluxqueue-scheduler:latest \
  --set postgres.image=${REGISTRY}/fluxqueue-postgres:latest \
  --set controllerManager.manager.imagePullPolicy=Never \
  --namespace fluxqueue-system \
  --create-namespace \
  --set scheduler.pullPolicy=Never \
  --set postgres.pullPolicy=Never \
        fluxqueue chart/
