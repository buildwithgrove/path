#!/bin/bash

# To experiment with this script and make them available in your shell, run:
# source ./e2e/scripts/shannon_preliminary_services_helpers.sh

#!/bin/bash

# =====================
# HELP COMMAND for shannon_preliminary_services_helpers.sh
# =====================
function help() {
  echo "=========================================="
  echo "Shannon Preliminary Services Query Utilities"
  echo "=========================================="
  echo ""
  echo "Available commands:"
  echo "  shannon_query_services_by_owner      - List all services owned by a given address"
  echo "  shannon_query_service_tlds_by_id     - Aggregate service endpoint TLDs by service ID"
  echo ""
  echo "Quick start examples:"
  echo "  shannon_query_services_by_owner main"
  echo "  shannon_query_services_by_owner main pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s"
  echo "  shannon_query_service_tlds_by_id main"
  echo "  shannon_query_service_tlds_by_id main --structured"
  echo "  shannon_query_service_tlds_by_id main --service-id eth"
  echo ""
  echo "====================================================="
  echo "Use --help with any command for detailed information"
  echo "====================================================="
  echo ""
  echo "NETWORK ENVIRONMENTS:"
  echo "  alpha, beta, main - Pocket Network Shannon environments"
  echo ""
  echo "TIPS:"
  echo "  You can inspect available service and supplier fields in the raw JSON files:"
  echo "    /tmp/shannon_all_services_<env>.json     (created by shannon_query_services_by_owner)"
  echo "    /tmp/shannon_supplier_dump_<env>.json    (created by shannon_query_service_tlds_by_id)"
  echo ""
  echo "  Default owner address: pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s"
  echo "  Alternative owner:     pokt100ea839pz5e9zuhtjxvtyyzuv4evhmq95682zw"
  echo ""
  echo "Requires 'pocketd' CLI to be installed and configured with ~/.pocket home directory"
}
help

# Run with --help for usage info.
function shannon_query_services_by_owner() {
  if [[ -z "$1" || "$1" == "--help" || "$1" == "-h" ]]; then
    echo "shannon_query_services_by_owner: List all services owned by a given address."
    echo ""
    echo "EXAMPLES:"
    echo "  # List all services for the default owner in mainnet:"
    echo "  shannon_query_services_by_owner main"
    echo ""
    echo "  # List all services for a specific owner address in beta for first owner"
    echo "  shannon_query_services_by_owner main pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s"
    echo ""
    echo "  # List all services for a specific owner address in beta for second owner"
    echo "  shannon_query_services_by_owner main pokt100ea839pz5e9zuhtjxvtyyzuv4evhmq95682zw"
    echo ""
    echo "USAGE:"
    echo "  shannon_query_services_by_owner <env> [owner_address]"
    echo ""
    echo "ARGUMENTS:"
    echo "    <env>           Required. Network environment - must be one of: alpha, beta, main."
    echo "    [owner_address] Optional. Owner address to filter services by."
    echo "                    Defaults to pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s if not provided."
    echo ""
    echo "DESCRIPTION:"
    echo "  - Queries all services from the specified Pocket Network Shannon environment ('alpha', 'beta', or 'main')"
    echo "  - Filters them to show only services owned by a specific address"
    echo "  - Outputs a formatted list"
    echo ""
    echo "OUTPUT:"
    echo "  - Outputs a list of services owned by the specified address in the format:"
    echo "      - <service_id>: <service_name>"s
    echo "  - Raw query response is saved to /tmp/shannon_all_services_<env>.json."
    echo ""
    echo "SIDE EFFECTS:"
    echo "  - Creates/overwrites /tmp/shannon_all_services_<env>.json with the full service list."
    echo "  - Prints info and errors to standard output."
    echo ""
    return 0
  fi

  local ENV="$1"
  local OWNER="$2"
  if [[ -z "$OWNER" ]]; then
    OWNER="pokt1lf0kekv9zcv9v3wy4v6jx2wh7v4665s8e0sl9s"
    echo "No owner address provided, defaulting to $OWNER"
  fi

  local DUMP_FILE="/tmp/shannon_all_services_${ENV}.json"

  echo "Querying services from network: $ENV"
  if ! pocketd query service all-services --network="$ENV" --grpc-insecure=false -o json --page-limit=1000000 >"$DUMP_FILE"; then
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
    echo "shannon_query_service_tlds_by_id - Query and aggregate service endpoint TLDs by service ID from a Pocket Network Shannon environment."
    echo ""
    echo "EXAMPLES:"
    echo "  shannon_query_service_tlds_by_id main"
    echo "  shannon_query_service_tlds_by_id main --structured"
    echo "  shannon_query_service_tlds_by_id main --service-id eth"
    echo "  shannon_query_service_tlds_by_id main --service-id bsc --structured"
    echo "  shannon_query_service_tlds_by_id main --structured --service-id polygon"
    echo ""
    echo "USAGE:"
    echo "  shannon_query_service_tlds_by_id <env> [--structured] [--service-id <service_id>]"
    echo ""
    echo "DESCRIPTION:"
    echo "  - Queries all suppliers from the specified Shannon environment ('alpha', 'beta', or 'main')."
    echo "  - Aggregates the 2nd-level TLDs (e.g. 'nodefleet.net') for each service ID, and displays the results."
    echo "  - Used in shannon_preliminary_services_test.sh to populate the SERVICE_TLDS associative array."
    echo "  - Maps service IDs to their comma-separated TLD lists for reporting."
    echo ""
    echo "PARAMETERS:"
    echo "  <env>              Required. Network environment - must be one of: alpha, beta, main."
    echo "  [--structured]     Optional. Output raw JSON only (no log output); otherwise, human-readable output."
    echo "  [--service-id ID]  Optional. Filter results to only show TLDs for the specified service ID."
    echo ""
    echo "OUTPUT:"
    echo "  Without --structured: Human-readable list showing colored service IDs and their TLDs."
    echo "  With --structured:    Raw JSON object mapping service IDs to arrays of unique TLDs."
    echo "  With --service-id:    Results filtered to only the specified service ID."
    echo ""
    echo "SIDE EFFECTS:"
    echo "  - Writes raw JSON response to /tmp/shannon_supplier_dump_<env>.json for debugging."
    echo "  - Requires 'pocketd' CLI to be installed and configured."
    echo "  - Uses ~/.pocket as the home directory for pocketd."
    echo ""
    echo "EXAMPLE JSON OUTPUT (with --structured):"
    echo "  {\"eth\": [\"nodefleet.net\", \"grove.city\"], \"bsc\": [\"ankr.com\", \"quicknode.com\"]}"
    echo ""
    echo "EXAMPLE JSON OUTPUT (with --structured --service-id eth):"
    echo "  {\"eth\": [\"nodefleet.net\", \"grove.city\"]}"
    return 0
  fi

  # Parse arguments
  local ENV=""
  local STRUCTURED=""
  local SERVICE_ID=""

  while [[ $# -gt 0 ]]; do
    case $1 in
    --structured)
      STRUCTURED="--structured"
      shift
      ;;
    --service-id)
      SERVICE_ID="$2"
      shift 2
      ;;
    *)
      if [[ -z "$ENV" ]]; then
        ENV="$1"
      else
        echo "❌ Unexpected argument: $1"
        return 1
      fi
      shift
      ;;
    esac
  done

  # Validate required environment parameter
  if [[ -z "$ENV" ]]; then
    echo "❌ Environment parameter is required. Use one of: alpha, beta, main"
    return 1
  fi

  local HOME_DIR="$HOME/.pocket"
  local DUMP_FILE="/tmp/shannon_supplier_dump_${ENV}.json"

  if [[ "$STRUCTURED" != "--structured" ]]; then
    if [[ -n "$SERVICE_ID" ]]; then
      echo "Querying suppliers from network: $ENV (filtering for service ID: $SERVICE_ID)"
    else
      echo "Querying suppliers from network: $ENV"
    fi
  fi

  if ! pocketd query supplier list-suppliers --network="$ENV" --home="$HOME_DIR" --grpc-insecure=false -o json --page-limit=1000 --dehydrated >"$DUMP_FILE"; then
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

  # Build jq filter with optional service ID filtering
  local JQ_FILTER='
    reduce .supplier[] as $s (
      {};
      reduce ($s.services // [])[] as $svc (
        .;
        ($svc.endpoints // [] | map(
            .url
            | sub("^https?://"; "")
            | split("/")[0]
            | split(".")[-2:]
            | join(".")
          )
          | unique) as $tlds
        |
        .[$svc.service_id] += $tlds
      )
    )
    | with_entries(.value |= unique | .value |= sort)'

  # Add service ID filter if specified
  if [[ -n "$SERVICE_ID" ]]; then
    JQ_FILTER="$JQ_FILTER | {\"$SERVICE_ID\": .\"$SERVICE_ID\"} | with_entries(select(.value != null))"
  fi

  if [[ "$STRUCTURED" == "--structured" ]]; then
    jq "$JQ_FILTER" "$DUMP_FILE"
  else
    jq -r "$JQ_FILTER | to_entries[] | \"- \u001b[34m\(.key)\u001b[0m: \(.value)\"" "$DUMP_FILE"
  fi
}
