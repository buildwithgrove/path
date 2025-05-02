package framework

import (
	observations "github.com/buildwithgrove/path/observation/qos/framework"
)

// TODO_IN_THIS_PR: change all `*Kind*` enum names to `*Type*`.

func (s *Sanction) buildObservation() *observations.Sanction {
	return &observations.Sanction{
		Type:            translateToObservationSanctionType(s.Type),
		Reason:          s.Reason,
		ExpiryTimestamp: timestampProto(s.ExpiryTime),
	}
}

func buildSanctionFromObservation(obs *observations.Sanction) *Sanction {
	return &Sanction{
		Type:       translateFromObservationSanctionType(obs.GetType()),
		Reason:     obs.GetReason(),
		ExpiryTime: timeFromProto(obs.GetExpiryTimestamp()),
	}
}

// DEV_NOTE: you MUST update this function when changing the set of valid sanction types.
func translateToObservationSanctionType(sanctionType SanctionType) observations.SanctionType {
	switch sanctionType {
	case SanctionTypeTemporary:
		return observations.SanctionType_SANCTION_TYPE_TEMPORARY
	case SanctionTypePermanent:
		return observations.SanctionType_SANCTION_TYPE_PERMANENT
	default:
		return observations.SanctionType_SANCTION_TYPE_UNSPECIFIED
	}
}

// DEV_NOTE: you MUST update this function when changing the set of valid sanction types.
func translateFromObservationSanctionType(sanctionType observations.SanctionType) SanctionType {
	switch sanctionType {
	case observations.SanctionType_SANCTION_TYPE_TEMPORARY:
		return SanctionTypeTemporary
	case observations.SanctionType_SANCTION_TYPE_PERMANENT:
		return SanctionTypePermanent
	default:
		return SanctionTypeUnspecified
	}
}
