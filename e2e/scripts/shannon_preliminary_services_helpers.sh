#!/bin/bash

# shannon_query_services_by_owner queries all services from a Pocket Network Shannon environment
# and filters them to show only services owned by a specific address.
#
# Parameters:
#   $1 (env): Network environment - must be one of: alpha, beta, main
#   $2 (owner_address): The owner address to filter services by
#
# Returns:
#   Outputs a list of services owned by the specified address in the format:
#   "- <service_id>: <service_name>"
#   Services are displayed with colored service IDs (blue)
#
# Side effects:
#   - Writes raw JSON response to /tmp/shannon_all_services_${ENV}.json for debugging
#   - Requires pocketd CLI tool to be installed and configured
#   - Uses ~/.pocket_prod as the home directory for pocketd
#
# Usage:
#   shannon_query_services_by_owner main pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s
function shannon_query_services_by_owner() {
  if [[ -z "$1" || "$1" == "--help" || "$1" == "-h" ]]; then
    echo "shannon_query_services_by_owner - List services owned by a given address"
    echo ""
    echo "USAGE:"
    echo "  shannon_query_services_by_owner <env> <owner_address>"
    echo ""
    echo "EXAMPLES:"
    echo "  shannon_query_services_by_owner main pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s"
    return 0
  fi

  local ENV="$1"
  local OWNER="$2"
  local HOME_DIR="$HOME/.pocket_prod"
  local DUMP_FILE="/tmp/shannon_all_services_${ENV}.json"

  echo "Querying services from network: $ENV"
  if ! pocketd query service all-services --network="$ENV" --home="$HOME_DIR" --grpc-insecure=false -o json >"$DUMP_FILE"; then
    echo "❌ Failed to query service list"
    return 1
  fi

  echo "✅ Response written to $DUMP_FILE"
  echo "🔍 Filtering services owned by: $OWNER"

  jq -r --arg owner "$OWNER" '
    .service[]
    | select(.owner_address == $owner)
    | "- \u001b[34m\(.id)\u001b[0m: \(.name)"
  ' "$DUMP_FILE"
}

# shannon_query_service_tlds_by_id queries all suppliers from a Pocket Network Shannon environment
# and aggregates the 2nd-level TLDs (e.g. 'nodefleet.net') for each service ID.
#
# This function is used in shannon_preliminary_services_test.sh to populate the SERVICE_TLDS
# associative array, which maps service IDs to their comma-separated TLD lists for display
# in the final report table.
#
# Parameters:
#   $1 (env): Network environment - must be one of: alpha, beta, main
#   $2 (structured): Optional "--structured" flag for JSON-only output without log messages
#
# Returns:
#   Without --structured: Human-readable list with colored service IDs showing TLDs per service
#   With --structured: Raw JSON object mapping service IDs to arrays of unique TLDs
#
# Side effects:
#   - Writes raw JSON response to /tmp/shannon_supplier_dump_${ENV}.json for debugging
#   - Requires pocketd CLI tool to be installed and configured
#   - Uses ~/.pocket as the home directory for pocketd
#
# Example JSON output with --structured:
#   {"eth": ["nodefleet.net", "grove.city"], "bsc": ["ankr.com", "quicknode.com"]}
#
# Usage:
#   shannon_query_service_tlds_by_id beta                    # Human-readable output
#   shannon_query_service_tlds_by_id main --structured      # JSON output for parsing
function shannon_query_service_tlds_by_id() {
  if [[ -z "$1" || "$1" == "--help" || "$1" == "-h" ]]; then
    echo "shannon_query_service_tlds_by_id - Aggregate service endpoint TLDs by service ID"
    echo ""
    echo "DESCRIPTION:"
    echo "  Queries the Pocket Network Shannon supplier list for a given network"
    echo "  and aggregates the 2nd-level TLDs (e.g. 'nodefleet.net') for each service ID."
    echo "  Also writes the raw JSON response to /tmp for debugging."
    echo ""
    echo "USAGE:"
    echo "  shannon_query_service_tlds_by_id <env> [--structured]"
    echo ""
    echo "ARGUMENTS:"
    echo "  env           Network environment - must be one of: alpha, beta, main"
    echo "  --structured  Output raw JSON only, no log output"
    echo ""
    echo "EXAMPLES:"
    echo "  shannon_query_service_tlds_by_id beta"
    echo "  shannon_query_service_tlds_by_id main --structured"
    return 0
  fi

  local ENV="$1"
  local STRUCTURED="$2"
  local HOME_DIR="$HOME/.pocket"
  local DUMP_FILE="/tmp/shannon_supplier_dump_${ENV}.json"

  if [[ "$STRUCTURED" != "--structured" ]]; then
    echo "Querying suppliers from network: $ENV"
  fi

  if ! pocketd query supplier list-suppliers --network="$ENV" --home="$HOME_DIR" --grpc-insecure=false -o json >"$DUMP_FILE"; then
    if [[ "$STRUCTURED" != "--structured" ]]; then
      echo "❌ Failed to query supplier list. Is the node running? Did you specify the right --home?"
      echo "Expected path: $HOME_DIR"
    fi
    return 1
  fi

  if [[ "$STRUCTURED" != "--structured" ]]; then
    echo "✅ Response written to $DUMP_FILE"
    echo "🔍 Parsing TLDs..."
  fi

  local JQ_FILTER='
      reduce .supplier[] as $s (
        {};
        . + (
          $s.services // []
          | map({
              key: .service_id,
              value: (
                .endpoints // []
                | map(
                    .url
                    | sub("^https?://"; "")
                    | split("/")[0]
                    | split(".")[-2:]
                    | join(".")
                  )
                | unique
              )
            })
          | from_entries
        )
      )
    '

  if [[ "$STRUCTURED" == "--structured" ]]; then
    jq "$JQ_FILTER" "$DUMP_FILE"
  else
    jq -r "$JQ_FILTER | to_entries[] | \"- \u001b[34m\(.key)\u001b[0m: \(.value)\"" "$DUMP_FILE"
  fi
}
