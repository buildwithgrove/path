---
sidebar_position: 9
title: HTTP Status Codes
description: Opinionated Approach to HTTP Status CodesDetails on running PATH locally with various configurations
---

:::danger DOCUMENTATION IN FLUX

**🦖 This documentation is very dynamic and in flux. Should be treated as a WIP.**

TODO_DOCUMENT(@olshansk): Update this page and move it to the right location

:::

## Opinionated Stance on HTTP Status Codes

There is no official mapping of JSON-RPC error codes to HTTP status codes in the [JSON-RPC 2.0 specification](https://www.jsonrpc.org/specification). JSON-RPC 2.0 is transport-agnostic and does not mandate any particular use of HTTP status codes.

PATH takes an opinionated stance on mapping JSON-RPC errors to HTTP status codes. This is a common practice but not an industry standard.

## JSON-RPC to HTTP Status Code Mapping

It’s common practice in JSON-RPC-over-HTTP implementations to map:
• `Client errors` (e.g., -32600 `Invalid Request`) to `4xx` HTTP statuses,
• `Server errors` (e.g., -32603 `Internal error` or -32000 “Server error”) to `5xx` HTTP statuses.

PATH follows this practice and maps JSON-RPC errors to HTTP status codes as follows:

| JSON-RPC Error Code           | Common Meaning        | Recommended HTTP Status |
| ----------------------------- | --------------------- | ----------------------- |
| **-32700**                    | Parse error           | **400**                 |
| **-32600**                    | Invalid request       | **400**                 |
| **-32601**                    | Method not found      | **404**                 |
| **-32602**                    | Invalid params        | **400** or **422**      |
| **-32603**, **-32000…-32099** | Internal/Server error | **500**                 |

## Zero Status Code

As explained [in this StackOverflow](https://stackoverflow.com/a/19862540/768439), if
a concrete HTTP status code cannot be determined, `0` is returned to the user.

**This is an opinionated approach and may be revisited in the future.**

## References

- Original GitHub Issue can be found [here](https://github.com/buildwithgrove/path/issues/179)
- Original GitHub PR can be found [here](https://github.com/buildwithgrove/path/pull/186/files)
