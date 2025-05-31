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

- **Protocol-level sanctions**: Endpoints that have been sanctioned due to errors or poor behavior
- **QoS-level disqualifications**: Endpoints failing quality-of-service checks

#### Why use it?

When developing or debugging PATH integrations, you may notice that certain endpoints aren't receiving traffic. This API helps you understand:

- Which endpoints are currently disqualified
- The specific reasons for disqualification
- Whether the disqualification is temporary (session-based) or permanent
- Aggregate statistics about endpoint health

#### How to use it

**Endpoint**: `GET /disqualified_endpoints`

**Required Headers**:
- `Target-Service-Id`: The service ID to query (e.g., `avax`, `eth`, `polygon`)

**Example Request**:
```bash
curl -X GET http://localhost:3001/disqualified_endpoints \
  -H "Target-Service-Id: avax"
```

#### Response Structure

The response contains three main sections:

##### 1. Protocol Level Data

Information about endpoints sanctioned by the protocol layer:

```json
{
  "protocol_level_data_response": {
    "permanently_sanctioned_endpoints": {
      // Map of permanently sanctioned endpoints
    },
    "session_sanctioned_endpoints": {
      // Map of temporarily sanctioned endpoints
      "node123abc:8f5b84bd49057:node456def:https://endpoint.example.com": {
        "endpoint_addr": "node123abc:8f5b84bd49057:node456def:https://endpoint.example.com",
        "reason": "relay error: relay: error sending request to endpoint",
        "service_id": "avax",
        "sanction_type": "SANCTION_SESSION",
        "error_type": "ENDPOINT_ERROR_TIMEOUT",
        "session_height": 12345,
        "created_at": "2023-01-15T10:41:37.94993+01:00"
      }
    },
    "permanent_sanctioned_endpoints_count": 0,
    "session_sanctioned_endpoints_count": 1,
    "total_sanctioned_endpoints_count": 1
  }
}
```

**Sanction Types**:
- `SANCTION_SESSION`: Temporary sanction for the current session
- `SANCTION_PERMANENT`: Permanent sanction until gateway restart

**Common Error Types**:
- `ENDPOINT_ERROR_TIMEOUT`: Request timeout
- `ENDPOINT_ERROR_CONNECTION`: Connection failure
- `ENDPOINT_ERROR_INVALID_RESPONSE`: Invalid or malformed response

##### 2. QoS Level Data

Information about endpoints failing quality-of-service checks:

```json
{
  "qos_level_data_response": {
    "disqualified_endpoints": {
      "node789xyz:a1b2c3d4e5f6:node012abc:https://rpc.example.com": {
        "endpoint_addr": "node789xyz:a1b2c3d4e5f6:node012abc:https://rpc.example.com",
        "reason": "invalid block number: endpoint returned block 12345, expected >= 12350",
        "service_id": "eth"
      }
    },
    "empty_response_count": 0,
    "chain_id_check_errors_count": 0,
    "archival_check_errors_count": 0,
    "block_number_check_errors_count": 1
  }
}
```

**QoS Check Types**:
- **Empty Response**: Endpoint returned no data
- **Chain ID Check**: Endpoint is on wrong chain
- **Archival Check**: Endpoint doesn't support historical queries
- **Block Number Check**: Endpoint is behind on block height

##### 3. Summary Statistics

Overall endpoint health metrics:

```json
{
  "total_service_endpoints_count": 10,
  "valid_service_endpoints_count": 8,
  "invalid_service_endpoints_count": 2
}
```

#### Common Use Cases

##### 1. Debugging Missing Endpoints

If you've added a new endpoint but it's not receiving traffic:

```bash
# Check if your endpoint is disqualified
curl -X GET http://localhost:3001/disqualified_endpoints \
  -H "Target-Service-Id: your-service-id" | jq '.'
```

##### 2. Monitoring Endpoint Health

Create a monitoring script to track endpoint health over time:

```bash
#!/bin/bash
while true; do
  echo "=== Endpoint Status at $(date) ==="
  curl -s http://localhost:3001/disqualified_endpoints \
    -H "Target-Service-Id: eth" | \
    jq '{
      total: .total_service_endpoints_count,
      valid: .valid_service_endpoints_count,
      invalid: .invalid_service_endpoints_count
    }'
  sleep 60
done
```

##### 3. Troubleshooting Specific Errors

When an endpoint is disqualified, check the reason to understand the issue:

- **Connection errors**: Check network connectivity and firewall rules
- **Timeout errors**: Endpoint may be overloaded or slow
- **Block height errors**: Endpoint may be out of sync
- **Chain ID errors**: Endpoint may be configured for wrong network

#### Best Practices

1. **Regular Monitoring**: Check disqualified endpoints regularly during development
2. **Automated Alerts**: Set up alerts when critical endpoints are disqualified
3. **Session vs Permanent**: Session sanctions clear on new sessions; permanent sanctions require gateway restart
4. **Root Cause Analysis**: Use the detailed error messages to fix underlying issues

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
