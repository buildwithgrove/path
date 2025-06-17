#!/bin/bash

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
