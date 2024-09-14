//go:build authorizer_plugin

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package postgres

import (
	"database/sql/driver"
	"fmt"
)

type RateLimitCapacityPeriod string

const (
	RateLimitCapacityPeriodDaily   RateLimitCapacityPeriod = "daily"
	RateLimitCapacityPeriodWeekly  RateLimitCapacityPeriod = "weekly"
	RateLimitCapacityPeriodMonthly RateLimitCapacityPeriod = "monthly"
)

func (e *RateLimitCapacityPeriod) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = RateLimitCapacityPeriod(s)
	case string:
		*e = RateLimitCapacityPeriod(s)
	default:
		return fmt.Errorf("unsupported scan type for RateLimitCapacityPeriod: %T", src)
	}
	return nil
}

type NullRateLimitCapacityPeriod struct {
	RateLimitCapacityPeriod RateLimitCapacityPeriod `json:"rate_limit_capacity_period"`
	Valid                   bool                    `json:"valid"` // Valid is true if RateLimitCapacityPeriod is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRateLimitCapacityPeriod) Scan(value interface{}) error {
	if value == nil {
		ns.RateLimitCapacityPeriod, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RateLimitCapacityPeriod.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRateLimitCapacityPeriod) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RateLimitCapacityPeriod), nil
}
