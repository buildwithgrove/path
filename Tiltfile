# Load necessary Tilt extensions
load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# Note: this list needs to be updated each time a new package is added to the repo.
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

# Load the existing config file, if it exists, or use an empty dict as fallback
local_config_path = "local_config.yaml"
local_config = read_yaml(local_config_path, default={})

# PATH operation modes determine which services are loaded:
# 1. path_only - PATH Service Only
# 2. path_with_auth - PATH Service, External Auth Server, Envoy Proxy, PADS, Rate Limiter, Redis.
# The observability stack is loaded in both modes.
MODE = os.getenv("MODE", "path_with_auth")  # Default mode is "path_with_auth"

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

# Import configuration files into Kubernetes ConfigMap
configmap_create("path-config", from_file="local/path/config/.config.yaml", watch=True)

# Build an image with a path binary
docker_build_with_restart(
    "path",
    ".",
    dockerfile="Dockerfile",
    entrypoint="/app/path",
    live_update=[sync("bin/path", "/app/path")],
)

# Conditionally add port forwarding based on the mode
if MODE == "path_only":
    # Run PATH without any dependencies and port 3069 exposed
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
        port_forwards=["3069:3069"],
    )
else:
    # Run PATH with all dependencies and no port exposed
    # as all traffic must be routed through Envoy Proxy.
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
        resource_deps=[
            "ext-authz",
            "envoy-proxy",
            "path-auth-data-server",
            "ratelimit",
            "redis",
        ],
        port_forwards=["3069:3069"],
    )

if MODE == "path_with_auth":
    # ---------------------------------------------------------------------------- #
    #                             Envoy Auth Resources                             #
    # ---------------------------------------------------------------------------- #
    # 1. External Auth Server                                                      #
    # 2. Envoy Proxy                                                               #
    # 3. Path Auth Data Server (PADS)                                              #
    # 4. Rate Limiter                                                              #
    # 5. Redis                                                                     #
    # ---------------------------------------------------------------------------- #

    # Import Envoy Auth configuration file into Kubernetes ConfigMaps
    configmap_create(
        "envoy-config",
        from_file="./local/path/envoy/.envoy.yaml",
        watch=True,
    )
    configmap_create(
        "allowed-services",
        from_file="./local/path/envoy/.allowed-services.lua",
        watch=True,
    )
    configmap_create(
        "gateway-endpoints",
        from_file="./local/path/envoy/.gateway-endpoints.yaml",
        watch=True,
    )
    configmap_create(
        "ratelimit-config",
        from_file="./local/path/envoy/.ratelimit.yaml",
        watch=True,
    )

    # 1. Build the External Auth Server image from envoy/auth_server/Dockerfile
    docker_build(
        "ext-authz",
        context="./envoy/auth_server",
        dockerfile="./envoy/auth_server/Dockerfile",
        live_update=[sync("./envoy/auth_server", "/app")],
    )
    # Load the Kubernetes YAML for the External Auth Server
    k8s_yaml("./local/kubernetes/envoy-ext-authz.yaml")
    k8s_resource(
        "ext-authz",
        labels=["envoy_auth"],
        port_forwards=["10003:10003"],
        resource_deps=["path-auth-data-server"],
    )

    # 2. Load the Kubernetes YAML for the envoy-proxy service
    k8s_yaml("./local/kubernetes/envoy-proxy.yaml")
    k8s_resource(
        "envoy-proxy",
        labels=["envoy_auth"],
        port_forwards=["3070:3070"],
    )

    # 3. Load the Kubernetes YAML for the path-auth-data-server service
    k8s_yaml("./local/kubernetes/envoy-pads.yaml")
    k8s_resource(
        "path-auth-data-server",
        labels=["envoy_auth"],
    )

    # 4. Load the Kubernetes YAML for the ratelimit service
    k8s_yaml("./local/kubernetes/envoy-ratelimit.yaml")
    k8s_resource(
        "ratelimit",
        labels=["envoy_auth"],
        resource_deps=["redis"],
    )

    # 5. Load the Kubernetes YAML for the redis service
    k8s_yaml("./local/kubernetes/envoy-redis.yaml")
    k8s_resource(
        "redis",
        labels=["envoy_auth"],
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
update_settings(k8s_upsert_timeout_secs=60)

helm_resource(
    "observability",
    "prometheus-community/kube-prometheus-stack",
    flags=[
        "--values=./local/kubernetes/observability-prometheus-stack.yaml",
        "--set=grafana.defaultDashboardsEnabled=true",
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

# TODO_UPNEXT(@adshmh): Define and import a custom Grafana dashboard.
# Use the poktroll Tiltfile as a template:
# https://github.com/pokt-network/poktroll/blob/12342f016f3238ee7840a85d5056b1fe5ada9767/Tiltfile#L157
