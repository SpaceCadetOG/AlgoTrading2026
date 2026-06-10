package runtime

import "strings"

func ProtectionForVenue(venue string) ProtectionCapabilities {
	venue = strings.ToLower(strings.TrimSpace(venue))
	switch venue {
	case "aster":
		return ProtectionCapabilities{
			Venue:                  venue,
			NativeStop:             true,
			NativeTakeProfit:       true,
			ReduceOnly:             true,
			RuntimeManagedTrailing: true,
			Ready:                  true,
			Reason:                 "native_stop_tp_reduce_only;trailing_runtime_managed",
		}
	case "hyperliquid":
		return ProtectionCapabilities{
			Venue:                    venue,
			ReduceOnly:               true,
			RuntimeManagedStop:       true,
			RuntimeManagedTakeProfit: true,
			RuntimeManagedTrailing:   true,
			Ready:                    true,
			Reason:                   "runtime_managed_stop_tp_trailing_reduce_only_exits",
		}
	case "lighter":
		return ProtectionCapabilities{
			Venue:                  venue,
			NativeStop:             true,
			NativeTakeProfit:       true,
			ReduceOnly:             true,
			RuntimeManagedTrailing: true,
			Ready:                  true,
			Reason:                 "documented_stop_loss_take_profit_reduce_only;trailing_runtime_managed",
		}
	default:
		return ProtectionCapabilities{Venue: venue, Reason: "unknown_venue_protection_capability"}
	}
}

func BuildProtectionPlan(candidate Candidate) ProtectionPlan {
	capability := ProtectionForVenue(candidate.Venue)
	plan := ProtectionPlan{
		Venue:         capability.Venue,
		StopArmed:     candidate.StopPrice > 0,
		TPLadderArmed: candidate.TP1 > 0 || candidate.TP2 > 0 || candidate.TP3 > 0,
		TrailingArmed: capability.RuntimeManagedTrailing,
		ReduceOnly:    capability.ReduceOnly,
		Reason:        capability.Reason,
	}
	if capability.NativeStop && capability.NativeTakeProfit {
		plan.Mode = "native_stop_tp_runtime_trailing"
	} else if capability.RuntimeManagedStop || capability.RuntimeManagedTakeProfit {
		plan.Mode = "runtime_managed_protection"
	} else {
		plan.Mode = "unsupported"
		plan.Reason = "protection_not_supported"
	}
	return plan
}
