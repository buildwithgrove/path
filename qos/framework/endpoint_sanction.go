package framework

import (
	"time"
)

// ======================
// Sanction Types
// ======================
// SanctionType identifies different types of endpoint sanctions.
type SanctionType int

const (
	SanctionTypeUnspecified SanctionType = iota // matches the "UNSPECIFIED" enum value in proto definitions.
	SanctionTypeTemporary                       // Time-limited exclusion
	SanctionTypePermanent                       // Permanent exclusion
)

// Sanction represents a recommendation to limit endpoint usage.
type Sanction struct {
	Type       SanctionType
	Reason     string
	ExpiryTime time.Time // Zero time means permanent
}
