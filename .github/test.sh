#!/bin/bash

set -eEu -o pipefail

# Keep track of root directory to return to
here=$(pwd)
registry=${1:-ghcr.io/converged-computing}
namespace=${2:-fluxqueue-system}

# These containers should already be loaded into minikube
echo "Sleeping 20 seconds waiting for images to deploy"
sleep 20
kubectl get pods -n ${namespace}

# Get pod for controller, scheduler, and postgres
controller_pod=$(kubectl get pods -n ${namespace} -o json | jq -r .items[0].metadata.name)
echo "Found fluxqueue controller pod: ${controller_pod}"
scheduler_pod=$(kubectl get pods -n ${namespace} -o json | jq -r .items[1].metadata.name)
echo "Found fluxqueue scheduler pod:  ${scheduler_pod}"
postgres_pod=$(kubectl get pods -n ${namespace} -o json | jq -r .items[2].metadata.name)
echo "Found fluxqueue postgres pod:   ${postgres_pod}"

# Wait for fluxion to pull (largest container)
while true
  do
  pod_status=$(kubectl get pods ${controller_pod} -n ${namespace} --no-headers -o custom-columns=":status.phase")
  if [[ "${pod_status}" == "Running" ]]; then
    echo "Controller ${controller_pod} is ready"
    break
  fi
  sleep 20
done

function echo_run {
  command="$@"
  echo "⭐️ ${command}"
  ${command}
}


# Show logs for debugging, if needed
echo
echo_run kubectl logs -n ${namespace} ${controller_pod} -c manager
echo
echo_run kubectl logs -n ${namespace} ${controller_pod} -c fluxion

echo
echo
echo_run kubectl logs -n ${namespace} ${scheduler_pod}

echo
echo
echo_run kubectl logs -n ${namespace} ${postgres_pod}

# Shared function to check output
function check_output {
  check_name="$1"
  actual="$2"
  expected="$3"
  if [[ "${expected}" != "${actual}" ]]; then
    echo "Expected output is ${expected}"
    echo "Actual output is ${actual}"
    exit 1
  fi
}

# Wait for webhook to be ready and submit the pod
while true
  do
  echo_run kubectl apply -f ./examples/pod.yaml
  retval=$?
  if [[ "${retval}" == "0" ]]; then
    echo "Webhook for ${controller_pod} is ready"
    break
  fi
  sleep 10
done

echo_run kubectl get pods

# The pod should be running, and scheduler should be fluxion
scheduled_by=$(kubectl get pod pod -o json | jq -r .spec.schedulerName)
pod_status=$(kubectl get pods pod --no-headers -o custom-columns=":status.phase")
echo
echo "                  Pod Status: ${pod_status}"
echo "                Scheduled by: ${scheduled_by}"
check_output 'check-pod-scheduled-by' "${scheduled_by}" "FluxionScheduler"
check_output 'check-pod-status' "${pod_status}" "Running"

# Now delete
echo_run kubectl delete -f ./examples/pod.yaml
sleep 2
pods_running=$(kubectl get pods -o json | jq -r '.items | length')
echo "                Pods Running: ${pods_running}"
check_output 'check-pod-deleted' "${pods_running}" "0"

# Do the same for a job
echo_run kubectl apply -f ./examples/job.yaml
sleep 3
echo_run kubectl get pods


# Check both job pods
for pod in $(kubectl get pods -o json | jq -r .items[].metadata.name)
  do 
  echo "Checking job pod ${pod}"
  scheduled_by=$(kubectl get pod ${pod} -o json | jq -r .spec.schedulerName)
  pod_status=$(kubectl get pods ${pod} --no-headers -o custom-columns=":status.phase")
  echo
  echo "                  Pod Status: ${pod_status}"
  echo "                Scheduled by: ${scheduled_by}"
  check_output 'check-pod-scheduled-by' "${scheduled_by}" "FluxionScheduler"
  check_output 'check-pod-status' "${pod_status}" "Running"
done


# Now delete the job - we are done!
echo_run kubectl delete -f ./examples/job.yaml
sleep 2
pods_running=$(kubectl get pods -o json | jq -r '.items | length')
echo "                Pods Running: ${pods_running}"
check_output 'check-pod-deleted' "${pods_running}" "0"