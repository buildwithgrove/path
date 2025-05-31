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

#### Response Structure

The response contains three main sections:

##### 1. Protocol Level Disqualified Endpoints

Information about endpoints sanctioned by the Shannon protocol layer:

```json
{
  "protocol_level_disqualified_endpoints": {
    "permanently_sanctioned_endpoints": {
      // Map of permanently sanctioned endpoints (keyed by endpoint URL)
    },
    "session_sanctioned_endpoints": {
      // Map of temporarily sanctioned endpoints (keyed by endpoint URL)
      "https://us-west-test-endpoint-1.demo": {
        "supplier_addresses": {
          "pokt13771d0a403a599ee4a3812321e2fabc509e7f3": {},
          "pokt183e1d77fc8a0a4a36f4deb5557553e55fe391c": {},
          // ... more supplier addresses
        },
        "endpoint_url": "https://us-west-test-endpoint-1.demo",
        "reason": "relay error: relay: error sending request to endpoint https://us-west-test-endpoint-1.demo: Post \"https://us-west-test-endpoint-1.demo\": dial tcp: lookup us-west-demo1-base-json.demo.do: no such host",
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
  }
}
```

**Key Fields**:
- `supplier_addresses`: Map of all supplier addresses using this endpoint URL
- `endpoint_url`: The sanctioned endpoint URL
- `reason`: Detailed error message explaining why the endpoint was sanctioned
- `session_id`: The session ID when the sanction was applied
- `app_addr`: The application address that triggered the sanction
- `sanction_type`: Type of sanction (SHANNON_SANCTION_SESSION or SHANNON_SANCTION_PERMANENT)
- `error_type`: Specific error type that caused the sanction
- `session_height`: Blockchain height when the session started
- `created_at`: Timestamp when the sanction was created

##### 2. QoS Level Disqualified Endpoints

Information about endpoints failing EVM-specific quality-of-service checks:

```json
{
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
  }
}
```

**QoS Check Types**:
- **Empty Response**: Endpoint returned no data
- **Chain ID Check**: Endpoint is on wrong chain (mismatched chain ID)
- **Archival Check**: Endpoint doesn't support historical/archival queries
- **Block Number Check**: Endpoint is behind on block height

##### 3. Summary Statistics

Overall endpoint health metrics:

```json
{
  "total_service_endpoints_count": 11,
  "qualified_service_endpoints_count": 9,
  "disqualified_service_endpoints_count": 2
}
```

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
