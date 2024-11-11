# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# Define directories for hot-reloading
hot_reload_dirs = [
    "./cmd",
    "./config",
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

# Define modes
MODE = os.getenv("MODE", "path_only")  # Default mode is "path_only"

# --------------------------------------------------------------------------- #
#                                PATH Service                                 #
# --------------------------------------------------------------------------- #

# Import PATH configuration file into Kubernetes ConfigMaps
configmap_create("path-config", from_file="./config/.config.yaml", watch=True)

# Build the PATH image from the Dockerfile in the root directory
docker_build(
    "path",
    context=".",
    dockerfile="./Dockerfile",
    live_update=[sync(".", "/app/path")],
)

# Load the Kubernetes YAML for the PATH service
k8s_yaml("./localnet/kubernetes/manifests/path.yaml")

# Conditionally add port forwarding based on the mode
if MODE == "path_only":
    k8s_resource(
        "path",
        labels=["path"],
        links=[link("http://localhost:3000", "Path Service")],
        port_forwards=["3000:3000"],
    )
else:
    k8s_resource(
        "path",
        labels=["path"],
        links=[link("http://localhost:3001", "Path Service via Envoy Proxy")],
    )

if MODE == "path_with_auth":
    # ----------------------------------------------------------------------------- #
    #                             Envoy Auth Containers                             #
    # ----------------------------------------------------------------------------- #

    # Import Envoy Auth configuration file into Kubernetes ConfigMaps
    configmap_create("envoy-config", from_file="./envoy/envoy.yaml", watch=True)
    configmap_create("gateway-endpoints", from_file="./envoy/gateway-endpoints.yaml", watch=True)
    configmap_create("ratelimit-config", from_file="./envoy/ratelimit.yaml", watch=True)

    # Import External Authorization Server environment variables into Kubernetes ConfigMaps
    configmap_create("ext-authz-env", from_env_file="./envoy/auth_server/.env", watch=True)

    # Build the External Authorization Server image from envoy/auth_server/Dockerfile
    docker_build(
        "ext-authz",
        context="./envoy/auth_server",
        dockerfile="./envoy/auth_server/Dockerfile",
        live_update=[sync("./envoy/auth_server", "/app")],
    )

    # Load the Kubernetes YAML for the External Authorization Server
    k8s_yaml("./localnet/kubernetes/manifests/ext-authz.yaml")
    k8s_resource(
        "ext-authz",
        labels=["envoy_auth"],
        port_forwards=["10003:10003"],
        links=[link("http://localhost:10003", "Ext Authz")],
        resource_deps=["path", "path-auth-data-server"],
    )

    # Load the Kubernetes YAML for the envoy-proxy service
    k8s_yaml("./localnet/kubernetes/manifests/envoy-proxy.yaml")
    k8s_resource(
        "envoy-proxy",
        labels=["envoy_auth"],
        port_forwards=["3001:3001"],
        links=[link("http://localhost:3001", "Envoy Proxy")],
        resource_deps=["path"],
    )

    # Load the Kubernetes YAML for the path-auth-data-server service
    k8s_yaml("./localnet/kubernetes/manifests/path-auth-data-server.yaml")
    k8s_resource(
        "path-auth-data-server",
        labels=["envoy_auth"],
        links=[link("http://localhost:50051", "Path Auth Data Server")],
        resource_deps=["path"],
    )

    # Load the Kubernetes YAML for the ratelimit service
    k8s_yaml("./localnet/kubernetes/manifests/ratelimit.yaml")
    k8s_resource(
        "ratelimit",
        labels=["envoy_auth"],
        links=[link("http://localhost:8081", "Ratelimit")],
        resource_deps=["path", "redis"],
    )

    # Load the Kubernetes YAML for the redis service
    k8s_yaml("./localnet/kubernetes/manifests/redis.yaml")
    k8s_resource(
        "redis",
        labels=["envoy_auth"],
        links=[link("http://localhost:6379", "Redis")],
        resource_deps=["path"],
    )

# ----------------------------------------------------------------------------- #
#                            Observability Resources                            #
# ----------------------------------------------------------------------------- #

helm_repo("prometheus-community", "https://prometheus-community.github.io/helm-charts")
helm_repo("grafana-helm-repo", "https://grafana.github.io/helm-charts")

# Increase timeout for building the image
update_settings(k8s_upsert_timeout_secs=60)

helm_resource(
    "observability",
    "prometheus-community/kube-prometheus-stack",
    flags=[
        "--values=./localnet/kubernetes/observability-prometheus-stack.yaml",
        "--set=grafana.defaultDashboardsEnabled=true",
    ],
    resource_deps=["prometheus-community"],
)

helm_resource(
    "loki",
    "grafana-helm-repo/loki-stack",
    flags=[
        "--values=./localnet/kubernetes/observability-loki-stack.yaml",
    ],
    labels=["monitoring"],
    resource_deps=["grafana-helm-repo"],
)

k8s_resource(
    new_name="grafana",
    workload="observability",
    extra_pod_selectors=[{"app.kubernetes.io/name": "grafana"}],
    port_forwards=["3003:3000"],
    labels=["monitoring"],
    links=[
        link("localhost:3003", "Grafana"),
    ],
    pod_readiness="wait",
    discovery_strategy="selectors-only",
)