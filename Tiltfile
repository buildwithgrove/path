# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# Note: this list needs to be updated each time a new package is added to the repo.
hot_reload_dirs = [
    "./local/path/.config.yaml",
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
    labels=["configuration"],
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

# ----------------------------------------------------------------------------------- #
#                                  PATH Resources                                     #
# ----------------------------------------------------------------------------------- #
# The following resources are installed from the PATH Helm chart.                     #
# Repo: https://github.com/buildwithgrove/helm-charts/tree/main/charts/path           #
# ----------------------------------------------------------------------------------- #
# 1. PATH (PATH API & Toolkit Harness): RPC/API Gateway                               #
# 2. GUARD (Gateway Utilities for Authentication, Routing & Defense): Envoy Gateway   #
# 3. WATCH (Workload Analytics and Telemetry for Comprehensive Health): Observability #
# ----------------------------------------------------------------------------------- #

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

# Tilt will run the Helm Chart with the following flags by default.
#
# For example:
# helm install path buildwithgrove/path \
#    --set config.fromSecret.enabled=true \
#    --set config.fromSecret.name=path-config \
#    --set config.fromSecret.key=.config.yaml
flags = [
# Enable PATH to load the config from a secret.
# PATH supports loading the config from either a Secret or a ConfigMap.
# See: https://github.com/buildwithgrove/helm-charts/blob/main/charts/path/values.yaml
    "--set", "config.fromSecret.enabled=true",
    "--set", "config.fromSecret.name=path-config",
    "--set", "config.fromSecret.key=.config.yaml",
]

# TODO_DOCUMENT(@commoddity): Add documentation for the .values.yaml file.
#
# Optional: Use a local values.yaml file to override the default values.
#
# For example, Tilt will append the flags:
#    --values ./local/path/.values.yaml --reset-values
# to the Helm command if the file exists.
#
# See file `./local/path/values.tmpl.yaml` for more details.
valuesFile = "./local/path/.values.yaml"
if read_yaml(valuesFile, default=None) != None:
    watch_file(valuesFile)
    flags.append("--reset-values") # Ensure that values are overridden by the .values.yaml file.
    flags.append("--values")
    flags.append(valuesFile)
    

# Run PATH Helm chart, including GUARD & WATCH.
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
    flags=flags,
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
# Uses a `k8s_resource` to displays logs for the `path` pod.
k8s_resource(
    workload="path",
    new_name="path",
    labels=["path"],
    port_forwards=["6060:6060"],
    extra_pod_selectors=[{"app.kubernetes.io/name": "path"}],
)

# 2. GUARD  Logs
# Uses a `local_resource` to display logs for the `envoy` and `gateway-helm` pods.
local_resource(
    "guard",
    cmd="echo 'Following GUARD logs...'",
    serve_cmd="kubectl logs -l app.kubernetes.io/name=envoy -l app.kubernetes.io/name=gateway-helm --follow",
    labels=["path"],
    resource_deps=["path"]
)

# 3. WATCH (Observability) Logs
# Uses a `local_resource` to display logs for the `grafana`, `kube-state-metrics`, and `prometheus-node-exporter` pods.
local_resource(
    "watch",
    cmd="echo 'Following WATCH logs...'", 
    serve_cmd="kubectl logs -l app.kubernetes.io/name=grafana -l app.kubernetes.io/name=kube-state-metrics -l app.kubernetes.io/name=prometheus-node-exporter --follow",
    labels=["path"],
    resource_deps=["path"]
)
