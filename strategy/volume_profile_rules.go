package strategy

func volumeProfileRulePack() []ExecutablePlaybook {
	risk := DefaultBookRiskRule()
	return []ExecutablePlaybook{
		{
			Name:          "VP Reversal + OF Absorption",
			DirectionBias: "both",
			Setup:         "Failed auction or sharp rejection at VAH, VAL, HVN, or POC with profile structure showing reversal context.",
			Entry: EntryRule{
				Name:       "Reversal Confirmation Entry",
				SourceBook: "volume_profile",
				EntryType:  "confirmation_retest",
				Trigger:    "Enter after rejection confirms and absorption, big limit order behavior, or cumulative delta divergence validates the reversal.",
				ConfirmationRequired: []string{
					"Rejection candle or failed auction confirmation",
					"Absorption, big limit order, or cumulative delta divergence",
				},
				InvalidIf: []string{
					"Price accepts beyond the failed level",
					"Order flow confirms the original move instead of the reversal",
				},
				OffsetBps: 5,
			},
			Stop: StopRule{
				Name:                   "Failed Auction Protective Stop",
				SourceBook:             "volume_profile",
				StopType:               "volume_based",
				Placement:              "Beyond the failed auction extreme or rejection wick, using the farthest meaningful barrier between profile edge and swing.",
				BufferBps:              8,
				CatastrophicMultiplier: 1.5,
			},
			Target: TargetRule{
				Name:        "Profile Reversal Ladder",
				SourceBook:  "volume_profile",
				TargetType:  "volume_profile_ladder",
				Target1:     "POC",
				Target2:     "Opposite value area boundary",
				Target3:     "Next HVN with trailing runner",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.50,
				PartialPct2: 0.30,
				PartialPct3: 0.20,
			},
			Management: TradeManagementRule{
				Name:           "Reversal Risk Compression",
				SourceBook:     "volume_profile",
				BreakEvenAfter: "TP1 or first clean reaction away from the failed level",
				TrailAfter:     "TP2 once reversal expands beyond value",
				EarlyExitIf: []string{
					"Price accepts beyond the failed level",
					"Follow-through dies immediately after entry",
				},
				ForceFlatIf: []string{
					"Funding hazard or macro hazard is near",
					"Opposite absorption appears before TP1",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Major profile level",
				"Failed auction or strong rejection",
				"Nearby POC, HVN, VAH, or VAL",
			},
			RequiredConfirmations: []string{
				"Absorption or big limit order",
				"Cumulative delta divergence or failed continuation read",
			},
			RejectIf: []string{
				"Reward-to-risk is below minimum",
				"Stop cannot be placed behind a clear failed level",
			},
		},
		{
			Name:          "VP Accumulation + OF Volume Cluster",
			DirectionBias: "both",
			Setup:         "Sideways accumulation zone forms with a POC or HVN inside the range and a breakout/retest context around a volume cluster.",
			Entry: EntryRule{
				Name:       "Accumulation Breakout Retest",
				SourceBook: "volume_profile",
				EntryType:  "breakout_retest",
				Trigger:    "Enter after breakout from accumulation and retest of POC, HVN, or cluster confirms hold from the new side.",
				ConfirmationRequired: []string{
					"Breakout from accumulation range",
					"Retest hold at POC, HVN, or volume cluster",
				},
				InvalidIf: []string{
					"Price accepts back into the accumulation range",
					"Retest slices through the level without reaction",
				},
				OffsetBps: 5,
			},
			Stop: StopRule{
				Name:                   "Accumulation Range Stop",
				SourceBook:             "volume_profile",
				StopType:               "combined_barrier",
				Placement:              "Beyond the opposite side of the accumulation range or behind VAL for longs and VAH for shorts, with a small basis-point buffer.",
				BufferBps:              10,
				CatastrophicMultiplier: 1.25,
			},
			Target: TargetRule{
				Name:        "Accumulation Expansion Targets",
				SourceBook:  "volume_profile",
				TargetType:  "profile_expansion",
				Target1:     "Nearest HVN or first profile barrier",
				Target2:     "Next major profile level",
				Target3:     "Trend runner if expansion persists",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.40,
				PartialPct2: 0.35,
				PartialPct3: 0.25,
			},
			Management: TradeManagementRule{
				Name:           "Accumulation Hold Management",
				SourceBook:     "volume_profile",
				BreakEvenAfter: "TP1",
				TrailAfter:     "New HVNs form in trend direction",
				EarlyExitIf: []string{
					"Price accepts back into the accumulation range",
					"Retest level loses cluster support",
				},
				ForceFlatIf: []string{
					"Range acceptance returns with no continuation",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Sideways accumulation zone",
				"Accumulation POC or HVN",
				"Volume cluster near the retest",
			},
			RequiredConfirmations: []string{
				"Breakout plus retest hold",
				"Volume cluster acting as support or resistance",
			},
			RejectIf: []string{
				"Profile level is unclear",
				"Reward-to-risk is below minimum",
			},
		},
		{
			Name:          "VP Trend + VWAP Pullback + Aggressive Delta",
			DirectionBias: "both",
			Setup:         "Trend profile is directional, VWAP is aligned with the move, and pullback reaches a profile level, VWAP, or prior acceptance zone.",
			Entry: EntryRule{
				Name:       "Trend Pullback Continuation",
				SourceBook: "volume_profile_vwap_orderflow",
				EntryType:  "pullback_continuation",
				Trigger:    "Enter after pullback to VWAP, deviation, or profile level holds and aggressive delta resumes in the trend direction.",
				ConfirmationRequired: []string{
					"Trend direction remains intact",
					"Aggressive delta resumes with trend",
					"VWAP or profile support/resistance reaction",
				},
				InvalidIf: []string{
					"Price accepts back into prior value",
					"Delta flips against continuation at the retest",
				},
				OffsetBps: 4,
			},
			Stop: StopRule{
				Name:                   "Trend Pullback Barrier Stop",
				SourceBook:             "vwap",
				StopType:               "combined_barrier",
				Placement:              "Beyond the pullback swing, VWAP barrier, or profile level that should hold, whichever is farthest and structurally cleanest.",
				BufferBps:              6,
				CatastrophicMultiplier: 1.25,
			},
			Target: TargetRule{
				Name:        "Trend Continuation Ladder",
				SourceBook:  "volume_profile_vwap",
				TargetType:  "hybrid_rr_profile",
				Target1:     "Fixed 1R",
				Target2:     "Next HVN or VWAP deviation",
				Target3:     "Trailing trend target",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.35,
				PartialPct2: 0.35,
				PartialPct3: 0.30,
			},
			Management: TradeManagementRule{
				Name:           "Trend Continuation Management",
				SourceBook:     "vwap",
				BreakEvenAfter: "TP1",
				TrailAfter:     "TP1 or TP2 when continuation proves itself",
				EarlyExitIf: []string{
					"Momentum fades",
					"Delta divergence appears against the trade",
					"Price accepts back into prior value",
				},
				ForceFlatIf: []string{
					"Funding hazard near decision zone",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Directional trend profile",
				"VWAP alignment",
				"Pullback into profile or VWAP support/resistance",
			},
			RequiredConfirmations: []string{
				"Aggressive delta with trend",
				"Continuation candle or reaction confirmation",
			},
			RejectIf: []string{
				"VWAP and profile disagree",
				"Reward-to-risk is below minimum",
			},
		},
		{
			Name:          "VP Rejection + Cumulative Delta Divergence",
			DirectionBias: "both",
			Setup:         "Price tests VAH, VAL, HVN, LVN, or POC and rejects, while cumulative delta diverges from price at the rejection point.",
			Entry: EntryRule{
				Name:       "Rejection Divergence Entry",
				SourceBook: "volume_profile_orderflow",
				EntryType:  "rejection_confirmation",
				Trigger:    "Enter after the rejection candle confirms and cumulative delta divergence shows failed participation in the tested direction.",
				ConfirmationRequired: []string{
					"Rejection candle confirmation",
					"Cumulative delta divergence",
				},
				InvalidIf: []string{
					"Price accepts past the rejected level",
					"Divergence disappears on the next push",
				},
				OffsetBps: 5,
			},
			Stop: StopRule{
				Name:                   "Rejection Wick Stop",
				SourceBook:             "volume_profile",
				StopType:               "volume_based",
				Placement:              "Beyond the rejection wick or failed auction extreme, with a buffer beyond the level that should hold.",
				BufferBps:              8,
				CatastrophicMultiplier: 1.5,
			},
			Target: TargetRule{
				Name:        "Rejection Rotation Targets",
				SourceBook:  "volume_profile",
				TargetType:  "rotation_targets",
				Target1:     "POC",
				Target2:     "Opposite value area boundary",
				Target3:     "Next HVN beyond value",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.45,
				PartialPct2: 0.35,
				PartialPct3: 0.20,
			},
			Management: TradeManagementRule{
				Name:           "Rejection Rotation Management",
				SourceBook:     "volume_profile",
				BreakEvenAfter: "TP1",
				TrailAfter:     "Only after TP2 if rotation expands",
				EarlyExitIf: []string{
					"Price accepts beyond the rejected level",
					"No reaction away from POC after entry",
				},
				ForceFlatIf: []string{
					"Macro or funding hazard is near while still in local value",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"VAH, VAL, HVN, LVN, or POC test",
				"Strong rejection",
			},
			RequiredConfirmations: []string{
				"Cumulative delta divergence",
				"Rejection candle or retest fail",
			},
			RejectIf: []string{
				"Reward-to-risk is below minimum",
				"Rejected level is not structurally clear",
			},
		},
	}
}
