# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# Note: this list needs to be updated each time a new package is added to the repo.
hot_reload_dirs = [
    "./local/path",
    "./local/guard",
    "./local/observability",
    "./cmd",
    "./config",
    "./gateway",
    "./health",
    "./message",
    "./qos",
    "./relayer",
    "./request",
    "./router",
]

# Load the existing config file, if it exists, or use an empty dict as fallback
local_config_path = "local_config.yaml"
local_config = read_yaml(local_config_path, default={})

# Configure helm chart reference.
# If using a local repo, set the path to the local repo; otherwise, use our own helm repo.
helm_repo(
    "buildwithgrove", 
    "https://buildwithgrove.github.io/helm-charts/",
    labels=["helm-charts"],
)
chart_prefix = "buildwithgrove/"
if local_config["helm_chart_local_repo"]["enabled"]:
    helm_chart_local_repo = local_config["helm_chart_local_repo"]["path"]
    hot_reload_dirs.append(helm_chart_local_repo)
    print("Using local helm chart repo " + helm_chart_local_repo)
    chart_prefix = helm_chart_local_repo + "/charts/"

# The folder containing the local configuration files.
LOCAL_DIR = "local"

# The folder containing PATH's local configuration files.
PATH_LOCAL_DIR = LOCAL_DIR + "/path"
# The configuration file for PATH.
PATH_LOCAL_CONFIG_FILE = PATH_LOCAL_DIR + "/.config.yaml"

# --------------------------------------------------------------------------- #
#                              Configuration Resources                        #
# --------------------------------------------------------------------------- #
# 1. PATH Config Updater                                                      #
# 2. Patch Envoy Gateway LoadBalancer                                         #
# --------------------------------------------------------------------------- #

# Start a Tilt resource to update the PATH config with the local config file.
local_resource(
    'path-config-updater',
    '''
    kubectl delete secret path-config --ignore-not-found=true && \
    kubectl create secret generic path-config --from-file=.config.yaml=./local/path/.config.yaml && \
    kubectl get deployment path > /dev/null 2>&1 && \
    kubectl rollout restart deployment path || \
    echo "Deployment not found - skipping rollout restart"
    ''',
    deps=[PATH_LOCAL_CONFIG_FILE],
    labels=["configuration"],
)

# Start a Tilt resource to patch the Envoy Gateway LoadBalancer resource 
# to ensure it is reachable from outside the cluster at "localhost:3070".
#
# For more context, see the comments at:
# `./local/scripts/patch_envoy_gateway.sh`.
local_resource(
    "patch-envoy-gateway",
    "./local/scripts/patch_envoy_gateway.sh",
    resource_deps=["path"],
    labels=["configuration"],
)
# --------------------------------------------------------------------------- #
#                              PATH Resources                                 #
# --------------------------------------------------------------------------- #
# The following resources are installed from a PATH Helm chart.               #
# 1. PATH                                                                     #
# 2. GUARD (Envoy Gateway)                                                    #
# 3. WATCH (Observability)                                                    #
# --------------------------------------------------------------------------- #

# TODO_TECHDEBT(@adshmh): use secrets for sensitive data with the following steps:
# 1. Add place-holder files for sensitive data
# 2. Add a secret per sensitive data item (e.g. gateway's private key)
# 3. Load the secrets into environment variables of an init container
# 4. Use an init container to run the scripts for updating config from environment variables.
# This can leverage the scripts under `e2e` package to be consistent with the CI workflow.

# Build an image with a PATH binary
docker_build_with_restart(
    "path",
    context=".",
    dockerfile="Dockerfile",
    entrypoint="/app/path",
    live_update=[
        sync("bin/path", "/app/path"),
        run("/app/path")
    ],
)

# Run PATH Helm chart, including:
# 1. PATH
# 2. GUARD (Envoy Gateway)
# 3. WATCH (Observability)
helm_resource(
    "path",
    chart_prefix + "path",
    image_deps=["path"],
    image_keys=[("image.repository", "image.tag")],
    links=[
        link(
            "http://localhost:3000/d/relays/path-service-requests?orgId=1",
            "Grafana dashboard",
        ),
    ],
    # Enable PATH to load the config from a secret.
    # PATH supports loading the config from either a Secret or a ConfigMap.
    # See: https://github.com/buildwithgrove/helm-charts/blob/main/charts/path/values.yaml
    flags=[
        "--set", "config.fromSecret.enabled=true",
        "--set", "config.fromSecret.name=path-config",
        "--set", "config.fromSecret.key=.config.yaml",
    ],
    # Port 6060 is exposed to serve pprof data.
    # Run the following commands to view the pprof data:
    #   $ make debug_goroutines
    port_forwards=["6060:6060"],
    resource_deps=["path-config-updater"]
)

# --------------------------------------------------------------------------- #
#                              Logs Resources                                 #
# --------------------------------------------------------------------------- #
# 1. PATH Logs                                                                #
# 2. GUARD (Envoy Gateway) Logs                                               #
# 3. WATCH (Observability) Logs                                               #
# --------------------------------------------------------------------------- #

# 1.PATH Logs
k8s_resource(
    workload="path",
    new_name="path",
    labels=["logs"],
    port_forwards=["6060:6060"],
    extra_pod_selectors=[{"app.kubernetes.io/name": "path"}],
)

# 2. GUARD (Envoy Gateway) Logs
local_resource(
    "guard",
    cmd="echo 'Following Envoy logs...'",  # A simple command that completes quickly
    serve_cmd="kubectl logs -l app.kubernetes.io/name=envoy -l app.kubernetes.io/name=gateway-helm --follow",
    labels=["logs"],
    resource_deps=["path"]
)

# 3. WATCH (Observability) Logs
local_resource(
    "watch",
    cmd="echo 'Following Kube State Metrics logs...'",  # A simple command that completes quickly
    # TODO_FIX_IN_THIS_PR(@commoddity): Fix the PVC issue that is stopping the Grafana pod from starting.
    # Then add -l app.kubernetes.io/name=grafana to the serve_cmd.
    serve_cmd="kubectl logs -l app.kubernetes.io/name=kube-state-metrics -l app.kubernetes.io/name=prometheus-node-exporter --follow",
    labels=["logs"],
    resource_deps=["path"]
)
