---
sidebar_position: 3
title: PATH Config
---

<div align="center">
<h1>PATH<br/>Path Configration YAML File</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>
</div>

:::info 

Envoy Proxy is configured with its own set of configuration files.

[For detailed information on the Envoy configuration, please refer to the Envoy Configuration Guide](../envoy/envoy_config.md).

:::

# Table of Contents <!-- omit in toc -->
- [Configuration YAML File](#configuration-yaml-file)
  - [Configuration File Location](#configuration-file-location)
  - [Example Configuration Files](#example-configuration-files)
  - [Config YAML Schema](#config-yaml-schema)
- [YAML Fields](#yaml-fields)
  - [`morse_config`](#morse_config)
  - [`shannon_config`](#shannon_config)
  - [`router_config`](#router_config)
  - [`hydrator_config`](#hydrator_config)
  - [`auth_server_config`](#auth_server_config)


## Configuration YAML File

All configuration for the PATH gateway is defined in a YAML file, which should be named `.config.yaml`.

```yaml
# Protocol Configuration
shannon_config:
  full_node_config:
    rpc_url: https://shannon-testnet-grove-rpc.beta.poktroll.com
    grpc_config:
      host_port: shannon-testnet-grove-grpc.beta.poktroll.com:443
  gateway_config:
    gateway_mode: "centralized"
    gateway_address: pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw
    gateway_private_key_hex: 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388
    owned_apps_private_keys_hex:
      - 40af4e7e1b311c76a573610fe115cd2adf1eeade709cd77ca31ad4472509d388

# Qos Hydrator Configuration
hydrator_config:
  service_ids:
    - "eth"
    - "solana"
    - "pokt"

# Auth Server Configuration
auth_server_config:
  grpc_host_port: path-auth-data-server:50051
  grpc_use_insecure_credentials: true
  endpoint_id_extractor_type: url_path
```

### Configuration File Location

The default location of the configuration file is `./config/.config.yaml` relative to the location of the PATH binary.

For example, when running the compiled PATH binary from the `./bin` directory, the configuration file will be located at `./bin/config/.config.yaml`.

As another example, when running PATH in Tilt, the configuration file is mounted in the container at `/app/config/.config.yaml`.

:::tip

The location of the configuration file may be overriden using the `-config` flag. 

For example, you may run`./path -config ./config/.config.custom.yaml`.

:::

### Example Configuration Files

Example configuration files for both Shannon and Morse gateways are provided below.

- [Example Shannon Config YAML File](https://github.com/buildwithgrove/path/blob/main/config/examples/config.shannon_example.yaml)
- [Example Morse Config YAML File](https://github.com/buildwithgrove/path/blob/main/config/examples/config.morse_example.yaml)

### Config YAML Schema

A YAML schema is provided for the configuration file. 

This schema is used to validate the configuration file and ensure that it is populated with the appropriate values.

- [Config YAML Schema File](https://github.com/buildwithgrove/path/tree/main/config/config.schema.yaml)

:::tip

For VSCode users, the [YAML Language Support by Red Hat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) plugin may be used to provide in-editor syntax highlighting and validation by installing the plugin and placing the following comment annotation at the top of your `.config.yaml` file:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/buildwithgrove/path/refs/heads/main/config/config.schema.yaml
```

:::

## YAML Fields

This is a comprehensive outline and explanation of each YAML field in the configuration file.
By default, the file must contain one (and only one) of the following top-level protocol-specific sections:

• `morse_config`  
• `shannon_config`  

All other sections are optional.

--------------------------------------------------------------------------------
### `morse_config` 
Required if operating the Gateway for the Morse protocol.

Fields within `morse_config`:

- `full_node_config` 
  - `url` (string, required): The URL of the `Pocket` node, which provides details about the current state of the network.
  - `relay_signing_key` (string, required): A 128-character hex-encoded private key used to sign Morse relays.  
    - Must be exactly 128 hex characters.  
  - `http_config` (optional): 
    - `retries` (integer): Number of retry attempts on HTTP requests.  
      - Defaults to 3 if omitted.
    - `timeout` (string): Duration of the HTTP request timeout, in Go duration format. 
      - Defaults to "5000ms" (5s) if omitted.

- `signed_aats` (required): 
  - This section contains `Application Authentication Token` (AAT) data for Morse. 
  - Each key in `signed_aats` must be a 40-character hex string representing the application address.  
  - For each entry (i.e., each appID/address key), the object must contain: 
    - `client_public_key` (string): 64-hex-character client public key.  
    - `application_public_key` (string): 64-hex-character application address public key.  
    - `application_signature` (string): 128-hex-character signature.  

--------------------------------------------------------------------------------
### `shannon_config` 
Required if operating the Gateway for the Shannon protocol.

Fields within `shannon_config`:

- `full_node_config` (required): 
  - `rpc_url` (string, required): The URL of the Shannon node’s RPC endpoint.  
  - `grpc_config` (required): 
    - `host_port` (string): Host and port for gRPC connections (e.g. "shannon-grpc.example.com:443")  
    - (Optional) Additional fields for backoff and keepalive behavior, which may be omitted to use defaults.  

- `gateway_config` (required): 
  - `gateway_mode` (string, required): The mode of the Shannon gateway. 
    - Must be one of "centralized", "delegated", or "permissionless".  
  - `gateway_address` (string, required): The gateway address in Bech32 format. 
    - Must match the pattern "^pokt1[0-9a-zA-Z]{38}$".  
  - `gateway_private_key_hex` (string, required): 64-hex-character private key.  
  - `owned_apps_private_keys_hex` (array of strings, required for "centralized" mode): A list of 64-hex-character private keys for Applications delegated to the Gateway.

--------------------------------------------------------------------------------
### `router_config`
Allows specifying server parameters for how the gateway handles incoming requests.

**All fields are optional.**

- `port` (integer): Port number on which the gateway server listens.
  - Defaults to 3069 if omitted.
- `max_request_body_size` (integer): Maximum request size in bytes.
  - Defaults to 1MB if omitted.
- `read_timeout` (string): Time limit for reading request data, in Go duration format.
  - Defaults to "5000ms" (5s) if omitted.
- `write_timeout` (string): Time limit for writing response data, in Go duration format.
  - Defaults to "10000ms" (10s) if omitted.
- `idle_timeout` (string): Time limit for closing idle connections, in Go duration format.
  - Defaults to "120000ms" (120s) if omitted.

--------------------------------------------------------------------------------
### `hydrator_config` 

To enable QoS for a service, the service ID must be provided here.

- `service_ids` (array of strings): Each string is a service ID for which hydrator logic may apply.

:::info

In order to enable QoS for a service, the ID provided here must match a service ID in [`config/service_qos.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos.go).

:::warning

Not all services currently have a QoS implementation in PATH; new QoS implementations are actively being worked on.

If a service ID is not present in [`config/service_qos.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos.go), a No-Op QoS implementation will be used for that service. This means a random endpoint will be selected for requests to that service.

:::


<!-- TODO_MVP: Add messaging_config -->

--------------------------------------------------------------------------------
### `auth_server_config` 
Used only by the External Auth Server. This is not used by the PATH Gateway logic itself, but is placed here for convenience.

- `grpc_host_port` (string): Host and port for the remote gRPC connection to the `Remote gRPC Server` (eg. PADS). 
  - Pattern requires a host:port format.
- `grpc_use_insecure_credentials` (boolean): Set to true if the `Remote gRPC Server` does not use TLS 
  - Defaults to false if omitted.  
- `endpoint_id_extractor_type` (string): Either "url_path" or "header". 
  - Specifies how endpoint IDs are extracted.
  - [See here for more details](../envoy/introduction.md#specifying-the-gateway-endpoint-id)
- `port` (integer): The local port for running the Auth Server
  - Defaults to 10003 if omitted.

:::info

[For detailed information on the External Auth Server, please refer to the External Auth Server Documentation](../envoy/introduction.md#external-auth-server).

:::
