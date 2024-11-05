load("ext://restart_process", "docker_build_with_restart")
load("ext://helm_resource", "helm_resource", "helm_repo")
load("ext://configmap", "configmap_create")

# A list of directories where changes trigger a hot-reload of PATH.
# Note: this list needs to be updated each time a new package is added to the repo.
hot_reload_dirs = ["cmd", "config", "gateway", "health", "message", "qos", "relayer", "request", "router"]

# Load the existing config file, if it exists, or use an empty dict as fallback
localnet_config_path = "localnet_config.yaml"
localnet_config = read_yaml(localnet_config_path, default={})

# TODO_UPNEXT(@adshmh): Package the default Helm chart for PATH and upload the pokt-network repo.
# Configure helm chart reference. If using a local repo, set the path to the local repo; otherwise, use our own helm repo.
helm_repo("pokt-network", "https://pokt-network.github.io/helm-charts/")
chart_prefix = "pokt-network/"
if localnet_config["helm_chart_local_repo"]["enabled"]:
    helm_chart_local_repo = localnet_config["helm_chart_local_repo"]["path"]
    hot_reload_dirs.append(helm_chart_local_repo)
    print("Using local helm chart repo " + helm_chart_local_repo)
    chart_prefix = helm_chart_local_repo + "/charts/"

# TODO_TECHDEBT(@adshmh): use secrets for sensitive data with the following steps:
# 1. Add place-holder files for sensitive data
# 2. Add a secret per sensitive data item (e.g. gateway's private key)
# 3. Load the secrets into environment variables of an init container
# 4. Use an init container to run the scripts for updating config from environment variables.
# This can leverage the scripts under `e2e` package to be consistent with the CI workflow.

# Import configuration files into Kubernetes ConfigMap
configmap_create("path-config", from_file="localnet/path/config/.config.yaml", watch=True)

# Build an image with a poktrolld binary
docker_build_with_restart(
    "path",
    ".",
    dockerfile="Dockerfile",
    entrypoint="/app/path",
    live_update=[sync("bin/path", "/app/path")],
)

# Provision PATH
helm_resource(
    "path",
    chart_prefix + "path",
    # TODO_UPNEXT(@adshmh): Add the CLI flag for loading the configuration file, once the CLI flags feature has been implemented.
    image_deps=["path"],
    image_keys=[("image.repository", "image.tag")],
)

# Observability
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

# TODO_UPNEXT(@adshmh): Define and import a custom Grafana dashboard.
# Use the poktroll Tiltfile as a template:
# https://github.com/pokt-network/poktroll/blob/12342f016f3238ee7840a85d5056b1fe5ada9767/Tiltfile#L157
