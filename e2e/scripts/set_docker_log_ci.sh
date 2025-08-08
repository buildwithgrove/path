#!/usr/bin/env bash
set -e
set -o nounset

# IMPORTANT: This script is only used for CI environments.
# It modifies the `e2e_load_test.config.tmpl.yaml` file
# and should never be run locally.

# This script sets the Docker log configuration for CI environments.
# It updates the E2E load test config to enable Docker logging to 
# stdout so we can capture Docker logs in the CI environment for debugging.

# Set the current working directory to e2e/config directory.
cd "$(dirname "$0")/../config" || exit 1

set_docker_log_ci() {
    local CONFIG_FILE="./e2e_load_test.config.tmpl.yaml"
    if [[ ! -f $CONFIG_FILE ]]; then
        echo "config file" $CONFIG_FILE "not found in" $PWD
        return 1
    fi

    # Update the config to enable Docker logging for CI
    yq -i '.e2e_load_test_config.e2e_config.docker_config.log_to_file = true' $CONFIG_FILE
    
    echo "Successfully set docker log_to_file to true in $CONFIG_FILE"
}

set_docker_log_ci "$@"
