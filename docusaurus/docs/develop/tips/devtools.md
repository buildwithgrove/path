---
sidebar_position: 1
title: DevTools
description: Developer Tools for Developing PATH
---

## DevTools

PATH provides developer tools to help diagnose and debug endpoint behavior in real-time.

### Disqualified Endpoints API

The `/disqualified_endpoints` endpoint is a powerful diagnostic tool that provides visibility into why certain endpoints are not being used by PATH for relay requests.

#### What is it?

The Disqualified Endpoints API returns a comprehensive list of endpoints that have been temporarily or permanently excluded from serving requests for a given service. This includes:

- **Protocol-level sanctions**: Endpoints sanctioned due to relay errors or poor behavior (managed by Shannon protocol)
- **QoS-level disqualifications**: Endpoints failing quality-of-service checks (managed by EVM QoS service)

#### Why use it?

When developing or debugging PATH integrations, you may notice that certain endpoints aren't receiving traffic. This API helps you understand:

- Which endpoints are currently disqualified and why
- Whether the disqualification is temporary (session-based) or permanent
- Aggregate statistics about endpoint health across your service
- Which suppliers are affected by sanctions

#### How to use it

**Endpoint**: `GET /disqualified_endpoints`

**Required Headers**:
- `Target-Service-Id`: The service ID to query (e.g., `base`, `eth`, `polygon`)

**Example Request**:
```bash
curl http://localhost:3069/disqualified_endpoints \
  -H "Target-Service-Id: base" | jq
```

Or use the provided make target:
```bash
make disqualified_endpoints SERVICE_ID=base
```

#### Response Structure

The response from the disqualified endpoints API contains details about endpoints that have been excluded from serving requests. Here's an example response:

```json
{
  "protocol_level_disqualified_endpoints": {
    "permanently_sanctioned_endpoints": {},
    "session_sanctioned_endpoints": {
      "pokt13771d0a403a599ee4a3812321e2fabc509e7f3-https://us-west-test-endpoint-1.demo": {
        "supplier_address": "pokt13771d0a403a599ee4a3812321e2fabc509e7f3",
        "endpoint_url": "https://us-west-test-endpoint-1.demo",
        "reason": "relay error: relay: error sending request to endpoint https://us-west-test-endpoint-1.demo: Post \"https://us-west-test-endpoint-1.demo\": dial tcp: lookup us-west-test-endpoint-1.demo: no such host",
        "service_id": "base",
        "session_id": "5a496c9faaabbaa1d184cf89ddfeb603ff515b990c6f714701b71572ab750ae8",
        "app_addr": "pokt1ccae0ce5ef5b1bcd74f3794f5b717b98a86412",
        "sanction_type": "SHANNON_SANCTION_SESSION",
        "error_type": "SHANNON_ENDPOINT_ERROR_TIMEOUT",
        "session_height": 23951,
        "created_at": "2025-05-31T14:57:41.484372+01:00"
      }
    },
    "permanent_sanctioned_endpoints_count": 0,
    "session_sanctioned_endpoints_count": 1,
    "total_sanctioned_endpoints_count": 1
  },
  "qos_level_disqualified_endpoints": {
    "disqualified_endpoints": {
      "pokt1ccae0ce5ef5b1bcd74f3794f5b717b98a86412-https://us-west-test-endpoint-1.demo": {
        "endpoint_addr": "pokt1ccae0ce5ef5b1bcd74f3794f5b717b98a86412-https://us-west-test-endpoint-1.demo",
        "reason": "endpoint has not returned an archival balance response to a \"eth_getBalance\" request",
        "service_id": "base"
      }
    },
    "empty_response_count": 0,
    "chain_id_check_errors_count": 0,
    "archival_check_errors_count": 1,
    "block_number_check_errors_count": 0
  },
  "total_service_endpoints_count": 11,
  "qualified_service_endpoints_count": 9,
  "disqualified_service_endpoints_count": 2
}
```

#### Field Descriptions

The response contains three main sections:

##### 1. Protocol Level Disqualified Endpoints

This section contains information about endpoints sanctioned by the Shannon protocol:

| Field                                  | Description                                                                       |
| -------------------------------------- | --------------------------------------------------------------------------------- |
| `permanently_sanctioned_endpoints`     | Map of endpoints with permanent sanctions (persist until gateway restart)         |
| `session_sanctioned_endpoints`         | Map of endpoints with temporary sanctions (expire after 1 hour or session change) |
| `permanent_sanctioned_endpoints_count` | Number of permanently sanctioned endpoints                                        |
| `session_sanctioned_endpoints_count`   | Number of temporarily sanctioned endpoints                                        |
| `total_sanctioned_endpoints_count`     | Total number of sanctioned endpoints                                              |

For each sanctioned endpoint:

| Field                | Description                                                               |
| -------------------- | ------------------------------------------------------------------------- |
| `supplier_addresses` | Map of supplier addresses using this endpoint URL                         |
| `endpoint_url`       | The URL of the sanctioned endpoint                                        |
| `reason`             | Detailed error message explaining the sanction reason                     |
| `service_id`         | The service ID the endpoint was serving                                   |
| `session_id`         | Session identifier when the sanction was applied                          |
| `app_addr`           | Application address that triggered the sanction                           |
| `sanction_type`      | Type of sanction (SHANNON_SANCTION_SESSION or SHANNON_SANCTION_PERMANENT) |
| `error_type`         | Specific error category (e.g., SHANNON_ENDPOINT_ERROR_TIMEOUT)            |
| `session_height`     | Blockchain height when the session started                                |
| `created_at`         | Timestamp when the sanction was created                                   |

##### 2. QoS Level Disqualified Endpoints

This section contains information about endpoints failing quality-of-service checks:

| Field                             | Description                                         |
| --------------------------------- | --------------------------------------------------- |
| `disqualified_endpoints`          | Map of endpoints failing QoS checks                 |
| `empty_response_count`            | Number of endpoints returning empty responses       |
| `chain_id_check_errors_count`     | Number of endpoints with incorrect chain ID         |
| `archival_check_errors_count`     | Number of endpoints failing historical data queries |
| `block_number_check_errors_count` | Number of endpoints with outdated block height      |

For each disqualified endpoint:

| Field           | Description                                                |
| --------------- | ---------------------------------------------------------- |
| `endpoint_addr` | Identifier for the endpoint (format: supplier-url)         |
| `reason`        | Detailed explanation of why the endpoint failed QoS checks |
| `service_id`    | The service ID the endpoint was serving                    |

##### 3. Summary Statistics

These fields provide an overview of endpoint health:

| Field                                  | Description                                         |
| -------------------------------------- | --------------------------------------------------- |
| `total_service_endpoints_count`        | Total number of endpoints for the requested service |
| `qualified_service_endpoints_count`    | Number of endpoints passing all checks              |
| `disqualified_service_endpoints_count` | Number of endpoints failing one or more checks      |

#### Implementation Details

The disqualified endpoints system has two main components:

1. **Protocol Level (Shannon)**:
   - Managed by `sanctionedEndpointsStore` in the Shannon protocol
   - Tracks both permanent and session-based sanctions
   - Session sanctions expire after 1 hour by default
   - Sanctions are applied based on relay errors and endpoint behavior

2. **QoS Level (EVM)**:
   - Managed by `serviceState` in the EVM QoS service
   - Performs synthetic checks: block number, chain ID, and archival support
   - Updates endpoint quality data based on responses
   - Filters out endpoints that don't meet quality requirements

#### Error Responses

**400 Bad Request** - Missing or invalid headers:
```json
{
  "error": "400 Bad Request",
  "message": "Target-Service-Id header is required"
}
```

**400 Bad Request** - Invalid service ID:
```json
{
  "error": "400 Bad Request",
  "message": "invalid service ID: no apps matched the request for service: invalid-service"
}
```
