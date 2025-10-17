# Grafana + VictoriaMetrics Export Guide

This guide shows how to use Grafana’s HTTP API with a **service account token** to export metrics and dashboards.
Tested on **Grafana v12.0.2**.

## Portal DB

The portal DB is the source of truth for running a SaaS using PATH to deploya service similar to [Grove's Portal](https://portal.grove.city)

See the following docs for more information:

- [portal-db](./portal-db/README.md)
- [portal-db/api](./portal-db/api/README.md)

---

## 🔑 1. Create a Service Account Token

1. Go to **Administration → Users and access → Service accounts**.
2. Click **New service account**, name it `metrics-exporter`, and assign role = **Viewer**.
3. Open the service account → click **Add service account token**.
4. Copy the token (looks like `glsa_XXXXXXXX...`).
   > ⚠️ Keep this secret — it allows read access to Grafana data sources.

---

## 📥 2. Export All Metric Names

```bash
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "https://grafana.tooling.grove.city/api/datasources/proxy/uid/ad44f9fa-1c08-4eae-ad3a-3a12d9fe762d/api/v1/label/__name__/values" \
  | jq -r '.data[]' > metrics.txt
```

This saves all metric names into `metrics.txt`.

---

## 📦 3. Export Metrics + Labels (Full Series)

```bash
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "https://grafana.tooling.grove.city/api/datasources/proxy/uid/ad44f9fa-1c08-4eae-ad3a-3a12d9fe762d/api/v1/series?match[]={__name__=~\".*\"}" \
  | jq -c '.data[]' > series.json
```

Each entry in `series.json` looks like:

```json
{ "__name__": "http_requests_total", "method": "GET", "status": "200" }
```

---

## 📊 4. Export a Single Dashboard (JSON)

Find the dashboard UID in its URL:
Example:

```
https://grafana.tooling.grove.city/d/bff5eb04-0c27-4cbd-ac27-d97b25530f5d/...
                           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ = UID
```

Export:

```bash
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "https://grafana.tooling.grove.city/api/dashboards/uid/bff5eb04-0c27-4cbd-ac27-d97b25530f5d" \
  | jq '.dashboard' > dashboard.json
```

---

## 📚 5. Export All Dashboards

List all dashboards:

```bash
curl -s \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "https://grafana.tooling.grove.city/api/search?query=" \
  | jq -r '.[].uid'
```

Loop through and save each:

```bash
for uid in $(curl -s -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  "https://grafana.tooling.grove.city/api/search?query=" | jq -r '.[].uid'); do
  echo "Exporting $uid..."
  curl -s -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    "https://grafana.tooling.grove.city/api/dashboards/uid/$uid" \
    | jq '.dashboard' > "dashboard-${uid}.json"
done
```

---

## ⚠️ Security Notes

- Treat your token like a password.
- Rotate/revoke tokens when not needed.
- Restrict tokens to **Viewer** role unless more is required.

---
