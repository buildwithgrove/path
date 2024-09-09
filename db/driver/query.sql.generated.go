// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package driver

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const selectUserApps = `-- name: SelectUserApps :many
SELECT u.id,
    u.account_id,
    u.secret_key,
    u.secret_key_required,
    COALESCE(
        jsonb_object_agg(
            w.type,
            jsonb_build_object(w.value, '{}'::jsonb)
        ) FILTER (
            WHERE w.user_app_id IS NOT NULL
        ),
        '{}'::jsonb
    )::jsonb AS allowlists,
    a.plan_type AS plan,
    p.rate_limit_throughput
FROM user_apps u
    LEFT JOIN user_app_allowlists w ON u.id = w.user_app_id
    LEFT JOIN accounts a ON u.account_id = a.id
    LEFT JOIN plans p ON a.plan_type = p.type
GROUP BY u.id,
    a.plan_type,
    p.rate_limit_throughput
`

type SelectUserAppsRow struct {
	ID                  string      `json:"id"`
	AccountID           pgtype.Text `json:"account_id"`
	SecretKey           pgtype.Text `json:"secret_key"`
	SecretKeyRequired   pgtype.Bool `json:"secret_key_required"`
	Allowlists          []byte      `json:"allowlists"`
	Plan                pgtype.Text `json:"plan"`
	RateLimitThroughput pgtype.Int4 `json:"rate_limit_throughput"`
}

func (q *Queries) SelectUserApps(ctx context.Context) ([]SelectUserAppsRow, error) {
	rows, err := q.db.Query(ctx, selectUserApps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SelectUserAppsRow
	for rows.Next() {
		var i SelectUserAppsRow
		if err := rows.Scan(
			&i.ID,
			&i.AccountID,
			&i.SecretKey,
			&i.SecretKeyRequired,
			&i.Allowlists,
			&i.Plan,
			&i.RateLimitThroughput,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
