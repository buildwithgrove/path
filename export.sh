#!/usr/bin/env bash
#
# Export Grafana metrics and dashboards
# Usage:
#   TOKEN=glsa_xxx ./export.sh
# or:
#   ./export.sh glsa_xxx

set -euo pipefail

# Token can come from ENV or first argument
TOKEN="${TOKEN:-${1:-}}"
if [[ -z "$TOKEN" ]]; then
  echo "❌ No token provided. Use TOKEN=... ./export.sh or ./export.sh TOKEN"
  exit 1
fi

GRAFANA_URL="https://grafana.tooling.grove.city"
DATASOURCE_UID="ad44f9fa-1c08-4eae-ad3a-3a12d9fe762d"

echo "📥 Exporting metric names..."
curl -s -H "Authorization: Bearer $TOKEN" \
  "$GRAFANA_URL/api/datasources/proxy/uid/$DATASOURCE_UID/api/v1/label/__name__/values" \
  | jq -r '.data[]' > metrics.txt
echo "✅ metrics.txt written"

# echo "📦 Exporting full series (metrics + labels)..."
# curl -s -H "Authorization: Bearer $TOKEN" \
#   "$GRAFANA_URL/api/datasources/proxy/uid/$DATASOURCE_UID/api/v1/series?match[]={__name__=~\".*\"}" \
#   | jq -c '.data[]' > series.json
# echo "✅ series.json written"

echo "📊 Exporting all dashboards..."
for uid in $(curl -s -H "Authorization: Bearer $TOKEN" \
  "$GRAFANA_URL/api/search?query=" | jq -r '.[].uid'); do
  echo "  - $uid"
  curl -s -H "Authorization: Bearer $TOKEN" \
    "$GRAFANA_URL/api/dashboards/uid/$uid" \
    | jq '.dashboard' > "dashboard-${uid}.json"
done
echo "✅ All dashboards exported"