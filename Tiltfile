# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# Note: this list needs to be updated each time a new package is added to the repo.
hot_reload_dirs = [
    "./cmd",
    "./config",
    "./local/path/config",
    "./gateway",
    "./health",
    "./message",
    "./qos",
    "./relayer",
    "./request",
    "./router",
    "./envoy",
    "./envoy/auth_server/auth",
    "./envoy/auth_server/endpoint_store",
    "./envoy/auth_server/proto",
]

# Load the existing config file, if it exists, or use an empty dict as fallback
local_config_path = "local_config.yaml"
local_config = read_yaml(local_config_path, default={})

# PATH operation modes determine which services are loaded:
#   1. (Default) 'path_with_auth' - PATH Service, External Auth Server, Envoy Proxy, PADS, Rate Limiter, Redis.
#   2. 'path_only' - PATH Service Only.
#   - The observability stack is loaded in both modes.

MODE = os.getenv("MODE", "path_with_auth")   # Default to 'path_with_auth' if MODE is not set

# Define the valid modes
VALID_MODES = ["path_only", "path_with_auth"]

# Check if the MODE is valid
if MODE not in VALID_MODES:
    fail("Invalid MODE: '{}'. Allowed values are {}. Please set a valid MODE.".format(MODE, VALID_MODES))
# --------------------------------------------------------------------------- #
#                                PATH Service                                 #
# --------------------------------------------------------------------------- #
# 1. PATH Service                                                             #
# --------------------------------------------------------------------------- #

# Configure helm chart reference.
# If using a local repo, set the path to the local repo; otherwise, use our own helm repo.
helm_repo("buildwithgrove", "https://buildwithgrove.github.io/helm-charts/")
chart_prefix = "buildwithgrove/"
if local_config["helm_chart_local_repo"]["enabled"]:
    helm_chart_local_repo = local_config["helm_chart_local_repo"]["path"]
    hot_reload_dirs.append(helm_chart_local_repo)
    print("Using local helm chart repo " + helm_chart_local_repo)
    chart_prefix = helm_chart_local_repo + "/charts/"

# TODO_TECHDEBT(@adshmh): use secrets for sensitive data with the following steps:
# 1. Add place-holder files for sensitive data
# 2. Add a secret per sensitive data item (e.g. gateway's private key)
# 3. Load the secrets into environment variables of an init container
# 4. Use an init container to run the scripts for updating config from environment variables.
# This can leverage the scripts under `e2e` package to be consistent with the CI workflow.\\

local_resource(
    'path-config-updater',
    '''
    kubectl delete secret path-config-local --ignore-not-found=true && \
    kubectl create secret generic path-config-local --from-file=.config.yaml=./local/path/config/.config.yaml && \
    kubectl get deployment path > /dev/null 2>&1 && \
    kubectl rollout restart deployment path || \
    echo "Deployment not found - skipping rollout restart"
    ''',
    deps=['./local/path/config/.config.yaml']
)

# Build an image with a path binary
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

# Port 6060 is exposed to serve pprof data.
# Run the following commands to view the pprof data:
#   $ make debug_goroutines
path_port_forwards = ["6060:6060"]

# Specify dependencies if PATH is running with auth.
# No ports (except 6060 for pprof), are exposed because all traffic MUST
# be routed through Envoy Proxy.
if MODE == "path_with_auth":
    path_resource_deps = [
        "guard",
    ]

# Specify the dependencies and port forwards if PATH is running WITHOUT auth.
if MODE == "path_only":
    # Run PATH without any dependencies
    path_resource_deps = ["path-config-updater"]
    # Expose port 3069 to serve relay requests (since envoy proxy is not used)
    path_port_forwards.append("3069:3069")

# Run PATH with dependencies and port forwarding settings matching the MODE:
#   1. With Auth: dependencies on envoy-proxy components, and NO exposed ports
#   2. Without Auth: no dependencies but exposing dedicated por
helm_resource(
    "path",
    chart_prefix + "path",
    flags=[
        "--values=./local/kubernetes/path-values.yaml",
    ],
    # TODO_MVP(@adshmh): Add the CLI flag for loading the configuration file.
    # This can only be done once the CLI flags feature has been implemented.
    image_deps=["path"],
    image_keys=[("image.repository", "image.tag")],
    labels=["path"],
    links=[
        link(
            "http://localhost:3000/d/relays/path-service-requests?orgId=1",
            "Grafana dashboard",
        ),
    ],
    port_forwards=path_port_forwards,
    resource_deps=path_resource_deps,
)

if MODE == "path_with_auth":
    # ---------------------------------------------------------------------------- #
    #                             Envoy Auth Resources                             #
    # ---------------------------------------------------------------------------- #
    # 1. Envoy Proxy (via Helm Chart)                                              #
    # 2. External Auth Server (via Helm Chart)                                     #
    # 3. Path Auth Data Server (PADS) (via Helm Chart)                             #
    # ---------------------------------------------------------------------------- #

    # New resources created from Helm Charts
    helm_resource(
        "guard",
        chart_prefix + "guard",
        labels=["guard"],
        port_forwards=["0.0.0.0:3070:3070"]
    )

# ----------------------------------------------------------------------------- #
#                            Observability Resources                            #
# ----------------------------------------------------------------------------- #
# 1. Prometheus                                                                 #
# 2. Loki                                                                       #
# 3. Grafana                                                                    #
# ----------------------------------------------------------------------------- #

helm_repo("prometheus-community", "https://prometheus-community.github.io/helm-charts")
helm_repo("grafana-helm-repo", "https://grafana.github.io/helm-charts")

# Increase timeout for building the image
update_settings(k8s_upsert_timeout_secs=120)

helm_resource(
    "observability",
    "prometheus-community/kube-prometheus-stack",
    flags=[
        "--values=./local/kubernetes/observability-prometheus-stack.yaml",
        "--set=grafana.defaultDashboardsEnabled="
        + str(local_config["observability"]["grafana"]["defaultDashboardsEnabled"]),
    ],
    resource_deps=["prometheus-community"],
)

helm_resource(
    "loki",
    "grafana-helm-repo/loki-stack",
    flags=[
        "--values=./local/kubernetes/observability-loki-stack.yaml",
    ],
    labels=["monitoring"],
    resource_deps=["grafana-helm-repo"],
)

k8s_resource(
    new_name="grafana",
    workload="observability",
    extra_pod_selectors=[{"app.kubernetes.io/name": "grafana"}],
    port_forwards=["3000:3000"],
    labels=["monitoring"],
    links=[
        link("localhost:3000", "Grafana"),
    ],
    pod_readiness="wait",
    discovery_strategy="selectors-only",
)

# Import custom grafana dashboards into Kubernetes ConfigMap
configmap_create("path-dashboards", from_file=listdir("local/grafana-dashboards/"))

# Grafana discovers dashboards to "import" via a label
local_resource(
    "path-dashboards-label",
    "kubectl label configmap path-dashboards grafana_dashboard=1 --overwrite",
    resource_deps=["path-dashboards"],
)
