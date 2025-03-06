#!/bin/bash
set -euo pipefail

# This script patches the Envoy Gateway LoadBalancer service to expose a static port (30070) inside the container.
# This allows Envoy Gateway to be reachable from the local machine on port 3070, as defined in the `kind-config.yaml` file.
#
# Context:
# - The Envoy Gateway service is created as a LoadBalancer which, when running in kind, automatically
#   uses a NodePort under the covers. 
# - In a real cloud environment, the LoadBalancer provisioner would map a public IP to this NodePort. 
# - In `kind`, since there's no external load balancer, Kubernetes auto-assigns a dynamic NodePort 
#   (within the 30000–32767 range) if one is not explicitly specified. 

PATH_NAMESPACE="path-local"
SERVICE_PREFIX="envoy-path-local-guard-envoy-gateway"
PATCH_PAYLOAD='[{"op": "replace", "path": "/spec/ports/0/nodePort", "value":30070}]'

# Gets the name of the Envoy Gateway service in the local cluster.
# eg. "envoy-path-local-guard-envoy-gateway-55375d27"
get_envoy_gateway_service_name() {
  local envoy_gateway_service_name
  envoy_gateway_service_name=$(kubectl get svc -n "$PATH_NAMESPACE" -o json | jq -r '.items[] | select(.metadata.name | startswith("'"$SERVICE_PREFIX"'")) | .metadata.name')
  echo "$envoy_gateway_service_name"
}

# Patches the Envoy Gateway service to enforce a consistent port number of 30070 inside the container.
# eg.  NAME                                           TYPE          CLUSTER-IP    EXTERNAL-IP  PORT(S)         AGE
#      envoy-path-local-guard-envoy-gateway-55375d27  LoadBalancer  10.96.50.102  <pending>    3070:30070/TCP  3m19s
patch_guard_port() {
    kubectl patch svc "$(get_envoy_gateway_service_name)" -n "$PATH_NAMESPACE" --type='json' -p="$PATCH_PAYLOAD"
}

echo "Waiting for Envoy Gateway service..."
while true; do
  svc=$(get_envoy_gateway_service_name)
  if [ -n "$svc" ]; then
    echo "Found Envoy Gateway service: $svc"
    echo "Patching service $svc..."
    patch_guard_port
    exit 0
  fi
  echo "Envoy Gateway service not found, retrying..."
  sleep 2
done