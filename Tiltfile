# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load('ext://k8s_attach', 'k8s_attach')
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# IMPORTANT_DEV_NOTE: this list needs to be updated each time a new package is added to the repo.
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
    "./protocol",
    "./metrics",
    "./observation",
    "./proto",
    "./websockets",
]

# merge_dicts updates the base dictionary with the updates dictionary.
def merge_dicts(base, updates):
    for k, v in updates.items():
        if k in base and type(base[k]) == "dict" and type(v) == "dict":
            # Assume nested dict and merge
            for vk, vv in v.items():
                base[k][vk] = vv
        else:
            # Replace or set the value
            base[k] = v

# Load the existing config file, if it exists, or use an empty dict as fallback
local_config_path = "local_config.yaml"
local_config_defaults = {
    "hot-reloading": True,
    "path": {
        "count": 1,
        "delve": {
            "enabled": False
        }
    },
    "observability": {
        "enabled": True,
        "grafana": {
            "defaultDashboardsEnabled": False
        }
    },
    # DEV_NOTE: to use a local copy of Helm charts, set BOTH of the following:
    #   1. enabled: true
    #   2. path: {PATH_TO_LOCAL_HELM_CHART}
    "helm_chart_local_repo": {
        "enabled": False,
        "path": "../helm-charts"
    }
}

# Initial empty config
local_config = {}
# Load the existing config file, if it exists, or use an empty dict as fallback
local_config_file = read_yaml(local_config_path, default={})
# Merge defaults into the local_config first
merge_dicts(local_config, local_config_defaults)
# Then merge file contents over defaults
merge_dicts(local_config, local_config_file)
# Check if there are differences or if the file doesn't exist
if (local_config_file != local_config) or (not os.path.exists(local_config_path)):
    print("Updating " + local_config_path + " with defaults")
    local("cat - > " + local_config_path, stdin=encode_yaml(local_config))

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
    chart_prefix = helm_chart_local_repo + "/charts/"
    # TODO_TECHDEBT(@okdas): Find a way to make this cleaner & performant w/ selective builds.
    local("cd " + chart_prefix + "guard && helm dependency update")
    local("cd " + chart_prefix + "path && helm dependency update")
    local("cd " + chart_prefix + "watch && helm dependency update")
    hot_reload_dirs.append(helm_chart_local_repo)
    print("Using local helm chart repo " + helm_chart_local_repo)


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
#
# TODO_TECHDEBT(@okdas): Remove this and the associated script once helm charts are updated.
local_resource(
    "patch-envoy-gateway",
    "./local/scripts/patch_envoy_gateway.sh",
    resource_deps=["path-stack"],
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

# if local_config["hot-reloading"]:
# Build the Go binary with proper settings for Alpine
local_resource(
    'path-binary',
    '''
    echo "Building Go binary..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o bin/path ./cmd 
    ''',
    deps=hot_reload_dirs,
    ignore=['**/node_modules', '.git'],
    labels=["hot-reloading"],
)

# Make sure path-binary runs before the Docker build
local_resource(
    "path-trigger",
    """
    echo "Triggering Docker build after binary build"
    touch .tilt-build-trigger
    """,
    resource_deps=["path-binary"],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=["hot-reloading"],
)

# Build an image with the PATH binary
docker_build_with_restart(
    "path-image",
    context=".",
    dockerfile_contents="""FROM golang:1.23.0
RUN apt-get -q update && apt-get install -qyy curl jq less
RUN mkdir -p /app/config
COPY bin/path /app/path
RUN chmod +x /app/path
WORKDIR /app
""",
    # only=["/app/path"],
    entrypoint=["/app/path"],
    live_update=[
        # First sync to a temporary location to avoid permission issues
        sync("bin/path", "/app/bin/path"),
        # Then run commands to properly handle the file
        run("cp -f /app/bin/path /app/path && chmod +x /app/path", trigger="bin/path")
    ],
    trigger='.tilt-build-trigger',  # Rebuild when this file changes
)

# Tilt will run the Helm Chart with the following flags by default.
#
# For example:
# helm install path buildwithgrove/path \
#    --set config.fromSecret.enabled=true \
#    --set config.fromSecret.name=path-config \
#    --set config.fromSecret.key=.config.yaml
flags = [
    # Enable GUARD resources.
    "--set", "guard.enabled=true",
    # Enable PATH to load the config from a secret.
    # PATH supports loading the config from either a Secret or a ConfigMap.
    # See: https://github.com/buildwithgrove/helm-charts/blob/main/charts/path/values.yaml
    "--set", "config.fromSecret.enabled=true",
    "--set", "config.fromSecret.name=path-config",
    "--set", "config.fromSecret.key=.config.yaml",
    # Always use the local image.
    "--set", "global.imagePullPolicy=Never",
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
    image_deps=["path-image"],
    image_keys=[("image.repository", "image.tag")],
    links=[
        link(
            # Forward port 3003 to Grafana's port 3000.
            # Port 3000 is already used by kind cluster's control plane.
            "http://localhost:3003/d/relays/path-service-requests?orgId=1",
            "Grafana dashboard",
        ),
    ],
    flags=flags,
    resource_deps=["path-config-updater"],
    labels=["path"],
)
update_settings(
    k8s_upsert_timeout_secs=90,
)

# --------------------------------------------------------------------------- #
#                              Logs Resources                                 #
# --------------------------------------------------------------------------- #
# 1. PATH Logs                                                                #
# 2. GUARD (Envoy Gateway) Logs                                               #
# 3. WATCH (Observability) Logs                                               #
# --------------------------------------------------------------------------- #

# 1.PATH Logs
# Uses a `k8s_resource` to display logs for the `path` pod.
k8s_resource(
    workload="path",
    new_name="path-stack",
    # Port 6060 is exposed to serve pprof data.
    # Run the following commands to view the pprof data:
    #   $ make debug_goroutines
    port_forwards=["6060:6060"],
    extra_pod_selectors=[{"app.kubernetes.io/name": "path"}],
    labels=["path"]
)

# Attach the proper port forwards to Grafana
# TODO_TECHDEBT(@okdas): Remove admin/password requirements.
k8s_attach(
    'path-grafana',
    'deployment/path-grafana',
    namespace='path',
    port_forwards="3003:3000",
    resource_deps=["path-stack"]
)

# 2. GUARD Logs - Waits for container readiness before following logs
local_resource(
    "guard-logs",
    cmd="echo 'Preparing to follow GUARD logs when pods are ready...'",
    serve_cmd='''
    echo "Waiting for GUARD pods to be fully ready..."
    until kubectl get pods -l 'app.kubernetes.io/name in (envoy,gateway-helm)' -o jsonpath='{.items[*].status.containerStatuses[*].ready}' 2>/dev/null | grep -q true; do
      echo "GUARD pods not ready yet..."; sleep 5
    done
    echo "GUARD pods ready, stabilizing..."; sleep 10
    echo "Following GUARD logs..."
    kubectl logs -l 'app.kubernetes.io/name in (envoy,gateway-helm)' --follow
    ''',
    labels=["k8s_logs"],
    resource_deps=["path-stack"]
)

# # 3. WATCH Logs - Waits for container readiness before following logs
# local_resource(
#     "watch-logs",
#     cmd="echo 'Preparing to follow WATCH logs when pods are ready...'",
#     serve_cmd='''
#     echo "Waiting for WATCH pods to be fully ready..."
#     until kubectl get pod -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].status.phase}' 2>/dev/null | grep -q Running &&
#           kubectl get pod -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null | grep -q true; do
#       sleep 5
#     done
#     echo "Checking other components..."
#     until kubectl get pods -l 'app.kubernetes.io/name in (kube-state-metrics,prometheus-node-exporter)' -o jsonpath='{.items[*].status.phase}' 2>/dev/null | tr ' ' '\n' | grep -v Running | wc -l | grep -q "^0$"; do
#       sleep 5
#     done
#     echo "All pods ready, stabilizing..."; sleep 20
#     echo "Following WATCH logs..."
#     kubectl logs -l 'app.kubernetes.io/name in (grafana,kube-state-metrics,prometheus-node-exporter)' --follow
#     ''',
#     labels=["k8s_logs"],
#     resource_deps=["path-stack"]
# )