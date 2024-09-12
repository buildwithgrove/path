<div align="center">
<h1>PATH<br/>Path API & Toolkit Harness</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

- [1. Introduction](#1-introduction)
  - [1.1 Prerequisites](#11-prerequisites)
- [2. Quickstart (Shannon)](#2-quickstart-shannon)
- [3. Configuration](#3-configuration)
  - [3.1 Configuration File](#31-configuration-file)
  - [3.2 Example Configuration Format](#32-example-configuration-format)
- [4. Running PATH](#4-running-path)
  - [4.1 Setup Config YAML](#41-setup-config-yaml)
  - [4.2 Start the Container](#42-start-the-container)
- [5. E2E Tests](#5-e2e-tests)
  - [5.1 Running Tests](#51-running-tests)
- [6. User Data](#6-user-data)
  - [6.1 Updated Endpoint](#61-updated-endpoint)
  - [6.2 Database Configuration](#62-database-configuration)
  - [6.3 Database Schema](#63-database-schema)

## 1. Introduction

**PATH** (Path API & Toolkit Harness) is an open source framework for enabling access to a decentralized supply network.

It provides various tools and libraries to streamline the integration and interaction with decentralized protocols.

### 1.1 Prerequisites

- [Docker](https://docs.docker.com/get-docker/)

**Required For Development:**

- [SQLC](https://docs.sqlc.dev/)
- [Mockgen](https://github.com/uber-go/mock)

## 2. Quickstart (Shannon)

1. Stake Apps and Gateway

   Refer to the [Poktroll Docker Compose Walkthrough](https://dev.poktroll.com/operate/quickstart/docker_compose_walkthrough) for instructions on staking your Application and Gateway on Shannon.

2. Populate Config File

   Run `make copy_config` to copy the example configuration file to `cmd/.config.yaml`.

   Update the configuration file `cmd/.config.yaml` with your Gateway's private key & address and your delegated Application's address.

   \*TIP: If you followed the [Debian Cheat Sheet](https://dev.poktroll.com/operate/quickstart/docker_compose_debian_cheatsheet#start-the-relayminer), you can run `path_prepare_config`
   to get you most of the way there. Make sure to review the `gateway_private_key` field.\*

3. Start the PATH Container

   Run `make path_up` to start the PATH container.

   _Example request to the PATH service:_

   ```bash
   curl http://eth-mainnet.localhost:3000/v1 \
       -X POST \
       -H "Content-Type: application/json" \
       -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
   ```

   _For detailed instructions on running PATH, see the [Running PATH](#running-path) section._

## 3. Configuration

### 3.1 Configuration File

The configuration for PATH is defined in a YAML file, which should be named `.config.yaml`.

This file is required for setting up a PATH instance and must be populated with the appropriate values.

The configuration is divided into several sections:

1. **Morse Configuration (`morse_config`)**:

   - **Required for Morse gateways.**
   - Must include full node URL and relay signing key.
   - Must include AAT data for all onchain staked applications assigned to the gateway operator

2. **Shannon Configuration (`shannon_config`)**:

   - **Required for Shannon gateways.**
   - Must include RPC URL, gRPC host/port, and gateway address/private key.
   - Must include the addresses of the onchain Applications that are delegated to the onchain Gateway.

3. **Services Configuration (`services`)**:

   - **Required for all gateways; at least one service must be listed.**
   - The key is the Service ID (e.g. `0021`) and the value is the service configuration.
   - Only the Service ID is required. All other fields are optional.

4. **Router Configuration (`router_config`)**:

   - _Optional. Default values will be used if not specified._
   - Configures router settings such as port and timeouts.

5. **User Data Configuration (`user_data_config`)**:
   - _Required only if the gateway operator wishes to associate user data with requests._
   - Configures the PostgreSQL database connection string.

### 3.2 Example Configuration Format

```yaml
shannon_config:
  full_node_config:
    rpc_url: "https://rpc-url.io"
    grpc_config:
      host_port: "grpc-url.io:443"
    gateway_address: "pokt1710ed9a8d0986d808e607c5815cc5a13f15dba"
    gateway_private_key: "d5fcbfb894059a21e914a2d6bf1508319ce2b1b8878f15aa0c1cdf883feb018d"
    delegated_app_addresses:
      - "pokt1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0"
      - "pokt1u2v3w4x5y6z7a8b9c0d1e2f3g4h5i6j7k8l9m0"

services:
  "0021":
    alias: "eth-mainnet"
```

- Full example config YAML files:
  - [Morse](./cmd/config/testdata/morse.example.yaml)
  - [Shannon](./cmd/config/testdata/shannon.example.yaml)
- [Config YAML Schema](./config/config.schema.yaml)

## 4. Running PATH

### 4.1 Setup Config YAML

- The PATH service requires the config YAML file to be populated.

1. The following `make` target is provided to copy the example configuration file - which is for the `Shannon` protocol - to the `.config.yaml` file needed by the PATH service:

   ```sh
   make copy_config
   ```

   **NOTE: For a full example of the config YAML format for both Shannon and Morse protocols, see the [example config YAML files](./cmd/config/testdata).**

2. You will then need to populate the `.config.yaml` file with the appropriate values for the protocol you wish to use.

   **⚠️ IMPORTANT: The data required to populate the `.config.yaml` file is sensitive and the contents of this file must never be shared outside of your organization. ⚠️**

### 4.2 Start the Container

1. Once the `.config.yaml` file is populated, to start the PATH service for a specific protocol, use the `make` target:

   ```sh
   make path_up
   ```

   **NOTE: The protocol version (`morse` or `shannon`) depends on whether `morse_config` or `shannon_config` is populated in the `.config.yaml` file.**

2. Once the Docker container is running, you may send service requests to the PATH service.

   By default, the PATH service will run on port `3000`.

3. To stop the PATH service, use the following `make` target:

   ```sh
   make path_down
   ```

## 5. E2E Tests

This repository contains end-to-end (E2E) tests for the Shannon relay protocol. The tests ensure that the protocol behaves as expected under various conditions.

To use E2E tests, a `make` target is provided to copy the example configuration file to the `.config.test.yaml` needed by the E2E tests:

```sh
make copy_test_config
```

Then update the `protocol.shannon_config.full_node_config` values with the appropriate values.

You can find the example configuration file [here](./e2e/.example.test.yaml).

Currently, the E2E tests are configured to run against the Shannon testnet.

Future work will include adding support for other protocols.

### 5.1 Running Tests

To run the tests, use the following `make` targets:

```sh
# Run all tests
make test_all

# Unit tests only
make test_unit

# Shannon E2E test only
make test_e2e_shannon_relay
```

## 6. User Data

By default, PATH does not associate user data with service requests (i.e. relays).

You may opt to enable user data config to unlock the ability to associate a specific user with a particular service request.

This is required for:

- User-specified app settings (e.g. secret key, etc.)
- Metering and billing of service requests (e.g. charging users $2 per 1 million requests)
- Rate limiting of service requests by throughput (e.g. 30 req / second) and/or capacity (e.g. 1M req / month)

### 6.1 Updated Endpoint

Enabling user data will modify the endpoint for service requests to require a user app identifier at the end of the URL path.

For example:

```bash
http://eth-mainnet.localhost:3000/v1/{user_app_id}
```

The default endpoint of `/v1` will no longer function without a user app ID.

### 6.2 Database Configuration

To enable user data, you must set up a Postgres database and populate the `.config.yaml` file's `user_data_config` field with the connection string.

```yaml
user_data_config:
  postgres_connection_string: "postgres://user:password@localhost:5432/database"
```

An example Postgres Docker configuration is included in the [docker-compose.yml](./docker-compose.yml) file at the root of this repository. **However, this configuration is not recommended for production use.**

### 6.3 Database Schema

[Base Schema SQL File](./db/driver/sqlc/schema.sql)

```mermaid
erDiagram
    PLANS {
        int id
        varchar type
        int rate_limit_throughput
        int rate_limit_capacity
        enum rate_limit_capacity_period
    }

    ACCOUNTS {
        varchar id
        varchar plan_type
    }

    USER_APPS {
        varchar id
        varchar account_id
        varchar secret_key
        boolean secret_key_required
    }

    PLANS ||--o{ ACCOUNTS : "plan_type"
    ACCOUNTS ||--o{ USER_APPS : "account_id"
```

A base schema is provided with the minimal tables and columns required to enable user data handling in PATH.

These tables should not be modified; instead, any additional functionality required by the gateway operator for managing user data should be added by extending the base tables and columns provided in this schema.

For example, it is up to the gateway operator to decide how they wish to manage their gateway's user data, user accounts, subscription plans, authentication, etc.
