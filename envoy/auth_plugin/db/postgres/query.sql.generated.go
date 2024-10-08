// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: query.sql

package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const selectGatewayEndpoints = `-- name: SelectGatewayEndpoints :many
SELECT ge.id,
    ge.account_id,
    ge.api_key,
    ge.api_key_required,
    ua.plan_type AS plan,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period
FROM gateway_endpoints ge
    LEFT JOIN user_accounts ua ON ge.account_id = ua.id
    LEFT JOIN plans p ON ua.plan_type = p.type
GROUP BY ge.id,
    ua.plan_type,
    p.rate_limit_throughput,
    p.rate_limit_capacity,
    p.rate_limit_capacity_period
`

type SelectGatewayEndpointsRow struct {
	ID                      string                      `json:"id"`
	AccountID               pgtype.Text                 `json:"account_id"`
	ApiKey                  pgtype.Text                 `json:"api_key"`
	ApiKeyRequired          pgtype.Bool                 `json:"api_key_required"`
	Plan                    pgtype.Text                 `json:"plan"`
	RateLimitThroughput     pgtype.Int4                 `json:"rate_limit_throughput"`
	RateLimitCapacity       pgtype.Int4                 `json:"rate_limit_capacity"`
	RateLimitCapacityPeriod NullRateLimitCapacityPeriod `json:"rate_limit_capacity_period"`
}

// This file is used by SQLC to autogenerate the Go code needed by the database driver.
// It contains all queries used for fetching user data by the Gateway.
// See: https://docs.sqlc.dev/en/latest/tutorials/getting-started-postgresql.html#schema-and-queries
func (q *Queries) SelectGatewayEndpoints(ctx context.Context) ([]SelectGatewayEndpointsRow, error) {
	rows, err := q.db.Query(ctx, selectGatewayEndpoints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SelectGatewayEndpointsRow
	for rows.Next() {
		var i SelectGatewayEndpointsRow
		if err := rows.Scan(
			&i.ID,
			&i.AccountID,
			&i.ApiKey,
			&i.ApiKeyRequired,
			&i.Plan,
			&i.RateLimitThroughput,
			&i.RateLimitCapacity,
			&i.RateLimitCapacityPeriod,
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
