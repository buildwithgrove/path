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
	_                     SanctionType = iota // skip the 0 value: it matches the "UNSPECIFIED" enum value in proto definitions.
	SanctionTypeTemporary              // Time-limited exclusion
	SanctionTypePermanent              // Permanent exclusion
)

// Sanction represents a recommendation to limit endpoint usage.
type Sanction struct {
	Type       SanctionType
	Reason     string
	ExpiryTime time.Time // Zero time means permanent
}
