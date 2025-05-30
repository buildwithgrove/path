---
sidebar_position: 3
title: How to Add Archival Configs
description: Description of how to add new EVM archival checks
---

## Table of Contents <!-- omit in toc -->

- [Overview](#overview)
- [Example - `Polygon zkEVM`](#example---polygon-zkevm)
  - [1. Go to the chain's block explorer](#1-go-to-the-chains-block-explorer)
  - [2. Choose a top account](#2-choose-a-top-account)
  - [3. Find the contract creation block](#3-find-the-contract-creation-block)
  - [4. Add the new archival check configuration](#4-add-the-new-archival-check-configuration)
  - [5. Send a test request](#5-send-a-test-request)


## Overview

<!-- TODO_MOVE(@commoddity): Update this section and merge it into JUDGE docs once JUDGE PR is merged -->

The process for adding new archival check configurations is somewhat manual but most be performed only once per chain.

Configuration must be added to [`service_qos_config.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos_config.go).

:::tip MORE INFO

For more information on the `service_qos_config.go` file, see the [Service QoS Config](./1_supported_services.md) page.

To learn more about how EVM archival checks work, see the [EVM Archival Checks](./2_evm_archival.md) page.

:::

## Example - `Polygon zkEVM`

This example uses the `Polygon zkEVM` chain (`F029`). Most EVM chain block explorers use a similar format for their browser UI.

- Block Explorer: https://zkevm.polygonscan.com/

### 1. Go to the chain's block explorer

   Go to the chain's block explorer and search for `Top Account` in the `Blockchain` dropdown.

   <div align="center">
   ![HowToArchival1](../../../static/img/howto_archival_1.png)
   </div>

### 2. Choose a top account

   Find an account with lots of activity and click on the `Address`.

   <div align="center">
   ![HowToArchival2](../../../static/img/howto_archival_2.png)
   </div>

### 3. Find the contract creation block

   Under the `Filters` section, select `View Contract Creation`.

   <div align="center">
   ![HowToArchival3](../../../static/img/howto_archival_3.png)
   </div>

   Take note of the block number of the first transaction for that address.

   <div align="center">
   ![HowToArchival4](../../../static/img/howto_archival_4.png)
   </div>

### 4. Add the new archival check configuration

   In the [`service_qos_config.go`](https://github.com/buildwithgrove/path/blob/main/config/service_qos_config.go) file, add a new entry to the `shannonServices` and/or `morseServices` array.

   :::important ARCHIVAL CONFIGURATION FORMAT

   The configuration must be entered in this exact format.

   ```go
   // Polygon zkEVM
   evm.NewEVMServiceQoSConfig("F029", "0x44d", evm.NewEVMArchivalCheckConfig(
      // https://zkevm.polygonscan.com/address/0xee1727f5074e747716637e1776b7f7c7133f16b1
      "0xee1727f5074E747716637e1776B7F7C7133f16b1",
      // Contract start block
      111,
   )),
   ```

   It must contain the following elements in `evm.NewEVMArchivalCheckConfig`, exactly as shown above.

   - Line 1: The URL for the contract address on the block explorer as a comment
      - _Example: `// https://zkevm.polygonscan.com/address/0xee1727f5074e747716637e1776b7f7c7133f16b1`_
   - Line 2: The contract address as the first parameter
      - _Example: `"0xee1727f5074E747716637e1776B7F7C7133f16b1"`_
   - Line 3: A comment containing `// Contract start block`
   - Line 4: A block number just slightly higher than the first transaction for that address as the second parameter
      - _Example: `111`_

   :::

### 5. Send a test request

   Configure PATH for the service you want to test, and run `make path_run` to start PATH from a local binary.

   :::tip 

   For information on how to configure PATH for a service, see one of the PATH cheatsheets:

   - [Shannon Cheat Sheet](../../develop/path/3_cheatsheet_shannon.md)
   - [Morse Cheat Sheet](../../develop/path/4_cheatsheet_morse.md)

   :::

   Then send a request to validate that data is returned correctly for the requested block.

   :::info EXAMPLE REQUEST

   Use an `eth_getBalance` request for:

   - The contract address
     - _Example: `0xee1727f5074E747716637e1776B7F7C7133f16b1`_
   - An old block hash, ideally close to the first transaction for that address
     - _Example: `0x15E` (350)_

   ```bash
   curl http://localhost:3069/v1 \
     -H "Target-Service-Id: F029" \
     -d '{
        "jsonrpc": "2.0",
        "method": "eth_getBalance",
        "id": 1,
        "params": [
           "0xee1727f5074E747716637e1776B7F7C7133f16b1",
           "0x15E"
        ]
     }'
   ```

   Response:

   ```json
   {
      "id": 1,
      "jsonrpc": "2.0",
      "result": "0x247a76d7647c0000"
   }
   ```

   :::
