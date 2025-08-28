#!/usr/bin/env bash
set -e
set -o nounset

# CI-ONLY: Modifies e2e_load_test.config.default.yaml for CI environments
# KEY FUNCTIONS:
# • Sets Docker log configuration for CI
# • Enables Docker logging to stdout
# • Captures Docker logs for CI debugging
# WARNING: Never run locally

# Set the current working directory to e2e/config directory.
cd "$(dirname "$0")/../../config" || exit 1

set_docker_log_ci() {
    local CONFIG_FILE="./e2e_load_test.config.default.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
        return 1
    fi

    # Update the config to enable Docker logging to stdout for CI
    yq -i '.e2e_load_test_config.e2e_config.docker_config.docker_log = true' $CONFIG_FILE

    echo "Successfully set docker_log to true in $CONFIG_FILE"
}

set_docker_log_ci "$@"
