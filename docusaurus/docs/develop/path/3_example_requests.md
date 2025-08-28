---
sidebar_position: 3
title: Example Relays
description: Example Relay requests to PATH
---

- [Env Setup](#env-setup)
- [Test Relay with `curl`](#test-relay-with-curl)
- [Test WebSockets with `wscat`](#test-websockets-with-wscat)
- [Load Testing Relays with `relay-util`](#load-testing-relays-with-relay-util)
- [Load Testing WebSockets with `websocket-load-test`](#load-testing-websockets-with-websocket-load-test)
- [Using `Portal App ID` instead of `API Key`](#using-portal-app-id-instead-of-api-key)

## Env Setup

Make sure you install optional tools first:

```bash
make install_tools_optional
```

## Test Relay with `curl`

Assuming you have an app staked for `eth`, you can query `eth_blockNumber`.

You can specify the service via the `Target-Service-Id` header:

```bash
curl http://localhost:3070/v1 \
 -H "Target-Service-Id: eth" \
 -H "Authorization: test_api_key" \
 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Or by using the `eth` subdomain:

```bash
curl http://eth.localhost:3070/v1 \
 -H "Authorization: test_api_key" \
 -d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'
```

Expected response:

```json
{ "id": 1, "jsonrpc": "2.0", "result": "0x2f01a" }
```

## Test WebSockets with `wscat`

:::tip

For `wscat` installation instructions, see [here](https://github.com/ArtiomL/wscat?tab=readme-ov-file#installation).

:::

```bash
wscat -c ws://localhost:3070/v1 \
 -H "Authorization: test_api_key" \
 -H "Target-Service-Id: xrplevm"
```

Expected terminal prompt:

```bash
Connected (press CTRL+C to quit)
>
```

And subscribe to events:

```bash
> {"jsonrpc":"2.0", "id": 1, "method": "eth_subscribe", "params": ["newHeads"]}
< {"jsonrpc":"2.0","result":"0x2dc4edb4ba815232ef2d144b5818c540","id":1}
```

Which will start sending events like so:

```bash
< {"jsonrpc":"2.0","method":"eth_subscription","params":{"subscription":"0x2dc4edb4ba815232ef2d144b5818c540","result":{"parentHash":"0xaf1ebef9181d53a61a05b328646e747b5100eaa7ea301e21f2b5b1772beda053", ...
```

:::info

This is a simple terminal-based WebSocket example and does not contain reconnection logic.

Connections will drop on session rollover, which is expected behavior.

In production environments, you should implement reconnection logic and handle errors gracefully.

:::

## Load Testing Relays with `relay-util`

You can use this helper to send 100 requests with performance metrics:

```bash
SERVICE_ID=eth make test_load__relay_util__local
```

Which runs the following command:

```bash
relay-util \
   -u http://localhost:3070/v1 \
   -H "target-service-id: ${SERVICE_ID}" \
   -H "Portal-Application-ID: ${GROVE_PORTAL_APP_ID}" \
   -H "authorization: ${GROVE_PORTAL_API_KEY}" \
   -d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
   -x 100 \
   -b
```

## Load Testing WebSockets with `websocket-load-test`

You can use this helper to subscribe to events:

```bash
SERVICE_ID=xrplevm make test_load__websocket_load_test__local
```

Subscribe to events:

```bash
websocket-load-test \
   --service "xrplevm" \
   --app-id $GROVE_PORTAL_APP_ID \
   --api-key $GROVE_PORTAL_API_KEY \
   --subs "newHeads,newPendingTransactions" \
   --count 10 \
   --log
```

## Using `Portal App ID` instead of `API Key`

For the requests above, if you're auth is configured to use Portal's authentication
instead of API keys, you can send a request like so:

```bash
curl http://eth.localhost:3070/v1/$GROVE_PORTAL_APP_ID \
 -X POST \
 -H "Authorization: $GROVE_PORTAL_API_KEY" \
 -H 'Content-Type: application/json' \
 -d '{ "method": "eth_blockNumber", "params": [], "id": 1, "jsonrpc": "2.0" }'
```

Or like so by sending the `Portal-Application-ID` header:

```bash
curl http://eth.localhost:3070/v1/ \
 -X POST \
 -H "Portal-Application-ID: $GROVE_PORTAL_APP_ID" \
 -H "Authorization: $GROVE_PORTAL_API_KEY" \
 -H 'Content-Type: application/json' \
 -d '{ "method": "eth_blockNumber", "params": [], "id": 1, "jsonrpc": "2.0" }'
```

For WebSockets, the equivalent would be:

```bash
wscat -c ws://localhost:3070/v1/$GROVE_PORTAL_APP_ID \
 -H "Authorization: $GROVE_PORTAL_API_KEY" \
 -H "Target-Service-Id: xrplevm"
```

Or, by sending the `Portal-Application-ID` header:

```bash
wscat -c ws://localhost:3070/v1 \
 -H "Portal-Application-ID: $GROVE_PORTAL_APP_ID" \
 -H "Authorization: $GROVE_PORTAL_API_KEY" \
 -H "Target-Service-Id: xrplevm"
```
