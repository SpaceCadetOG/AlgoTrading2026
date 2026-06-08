package strategy

func vwapRulePack() []ExecutablePlaybook {
	risk := DefaultBookRiskRule()
	return []ExecutablePlaybook{
		{
			Name:          "VWAP First Touch Combo",
			DirectionBias: "both",
			Setup:         "Price approaches VWAP from the correct side and the first touch lines up with a volume-profile level, order-flow confirmation, or price-action structure.",
			Entry: EntryRule{
				Name:       "VWAP First Touch Combo Entry",
				SourceBook: "vwap",
				EntryType:  "first_touch_combo",
				Trigger:    "Enter on the first VWAP touch only when confluence exists with profile or order-flow context.",
				ConfirmationRequired: []string{
					"VWAP approach from correct side",
					"Volume Profile or Order Flow confluence",
				},
				InvalidIf: []string{
					"No confluence at VWAP",
					"VWAP is touched in chop without reaction",
				},
				OffsetBps: 3,
			},
			Stop: StopRule{
				Name:                   "VWAP Combo Barrier Stop",
				SourceBook:             "vwap",
				StopType:               "combined_barrier",
				Placement:              "Beyond VWAP plus the recent swing or ATR-style structural buffer, choosing the farthest meaningful barrier.",
				BufferBps:              5,
				CatastrophicMultiplier: 1.25,
			},
			Target: TargetRule{
				Name:        "VWAP Combo Targets",
				SourceBook:  "vwap",
				TargetType:  "starter_plus_profile",
				Target1:     "1R starter target",
				Target2:     "Next VWAP deviation or profile barrier",
				Target3:     "Optional runner into broader structure",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.50,
				PartialPct2: 0.30,
				PartialPct3: 0.20,
			},
			Management: TradeManagementRule{
				Name:           "VWAP First Touch Management",
				SourceBook:     "vwap",
				BreakEvenAfter: "Clear favorable reaction or TP1",
				TrailAfter:     "Only after TP1 if move proves itself",
				EarlyExitIf: []string{
					"VWAP fails to react",
					"Confluence breaks immediately after touch",
				},
				ForceFlatIf: []string{
					"No confluence remains at the touch",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"VWAP approach from correct side",
				"At least one profile or order-flow confluence",
			},
			RequiredConfirmations: []string{
				"Confluence confirmation at touch",
			},
			RejectIf: []string{
				"First touch happens without confluence",
				"Reward-to-risk is below minimum",
			},
		},
		{
			Name:          "VWAP Successful Reaction",
			DirectionBias: "both",
			Setup:         "Price touches VWAP and reacts away, creating a safer entry than a blind first touch.",
			Entry: EntryRule{
				Name:       "VWAP Reaction Candle Entry",
				SourceBook: "vwap",
				EntryType:  "reaction_confirmation",
				Trigger:    "Enter after the VWAP touch reacts and a confirmation candle closes in the reaction direction.",
				ConfirmationRequired: []string{
					"VWAP touch",
					"Reaction candle in intended direction",
				},
				InvalidIf: []string{
					"Price accepts through VWAP after touch",
				},
				OffsetBps: 2,
			},
			Stop: StopRule{
				Name:                   "VWAP Reaction Stop",
				SourceBook:             "vwap",
				StopType:               "price_action_vwap",
				Placement:              "Beyond the reaction swing or just beyond VWAP if that is the level that should hold.",
				BufferBps:              4,
				CatastrophicMultiplier: 1.2,
			},
			Target: TargetRule{
				Name:        "VWAP Reaction Targets",
				SourceBook:  "vwap",
				TargetType:  "rr_then_structure",
				Target1:     "1R",
				Target2:     "VWAP first deviation",
				Target3:     "Next volume-profile barrier",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.50,
				PartialPct2: 0.30,
				PartialPct3: 0.20,
			},
			Management: TradeManagementRule{
				Name:           "VWAP Reaction Management",
				SourceBook:     "vwap",
				BreakEvenAfter: "Clear favorable follow-through or TP1",
				TrailAfter:     "After TP1 only",
				EarlyExitIf: []string{
					"Price accepts back through VWAP",
					"Reaction momentum disappears",
				},
				ForceFlatIf: []string{
					"VWAP loses relevance and price returns to chop",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"VWAP support or resistance reaction",
			},
			RequiredConfirmations: []string{
				"Price confirmation candle",
			},
			RejectIf: []string{
				"Stop is unclear",
				"Reward-to-risk is below minimum",
			},
		},
		{
			Name:          "VWAP Deviation Rotation",
			DirectionBias: "both",
			Setup:         "Price rotates between VWAP and the first deviation in a range, creating fade opportunities back toward fair value.",
			Entry: EntryRule{
				Name:       "Deviation Rotation Entry",
				SourceBook: "vwap",
				EntryType:  "deviation_rotation",
				Trigger:    "Fade the first deviation back toward VWAP after price reacts from the deviation instead of accepting beyond it.",
				ConfirmationRequired: []string{
					"Range or rotational market",
					"Reaction from first VWAP deviation",
				},
				InvalidIf: []string{
					"Price accepts beyond the deviation",
					"Trend conditions replace rotational conditions",
				},
				OffsetBps: 3,
			},
			Stop: StopRule{
				Name:                   "Deviation Extreme Stop",
				SourceBook:             "vwap",
				StopType:               "deviation_barrier",
				Placement:              "Beyond the deviation extreme that should not be accepted if rotation remains valid.",
				BufferBps:              5,
				CatastrophicMultiplier: 1.2,
			},
			Target: TargetRule{
				Name:        "Deviation Rotation Targets",
				SourceBook:  "vwap",
				TargetType:  "vwap_rotation",
				Target1:     "VWAP midpoint or line",
				Target2:     "Opposite first deviation",
				Target3:     "Optional full range extension if rotation persists",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.50,
				PartialPct2: 0.35,
				PartialPct3: 0.15,
			},
			Management: TradeManagementRule{
				Name:           "Deviation Rotation Management",
				SourceBook:     "vwap",
				BreakEvenAfter: "Return toward VWAP",
				TrailAfter:     "Only after VWAP is reclaimed and move continues",
				EarlyExitIf: []string{
					"Price accepts beyond deviation",
				},
				ForceFlatIf: []string{
					"Rotational context is replaced by a strong trend",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Rotational market between VWAP and first deviation",
			},
			RequiredConfirmations: []string{
				"Reaction from first deviation",
			},
			RejectIf: []string{
				"Trend conditions dominate",
				"Reward-to-risk is below minimum",
			},
		},
		{
			Name:          "VWAP Trend Deviation Continuation",
			DirectionBias: "both",
			Setup:         "Strong trend pulls back to the first VWAP deviation and uses that deviation as continuation support or resistance.",
			Entry: EntryRule{
				Name:       "Trend Deviation Continuation Entry",
				SourceBook: "vwap",
				EntryType:  "trend_deviation_continuation",
				Trigger:    "Enter after pullback to first deviation confirms continuation with supportive price action or aligned order-flow behavior.",
				ConfirmationRequired: []string{
					"Strong trend context",
					"Continuation confirmation from first deviation",
				},
				InvalidIf: []string{
					"Price accepts beyond the deviation",
					"Momentum fades or delta diverges against the move",
				},
				OffsetBps: 3,
			},
			Stop: StopRule{
				Name:                   "Trend Deviation Protective Stop",
				SourceBook:             "vwap",
				StopType:               "combined_barrier",
				Placement:              "Beyond the deviation and the pullback swing, using the strongest barrier among deviation, swing, profile level, and ATR-style fallback.",
				BufferBps:              6,
				CatastrophicMultiplier: 1.25,
			},
			Target: TargetRule{
				Name:        "Trend Deviation Continuation Targets",
				SourceBook:  "vwap",
				TargetType:  "rr_plus_extension",
				Target1:     "1R",
				Target2:     "Next deviation or extension target",
				Target3:     "Trailing trend target",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.40,
				PartialPct2: 0.30,
				PartialPct3: 0.30,
			},
			Management: TradeManagementRule{
				Name:           "Trend Deviation Management",
				SourceBook:     "vwap",
				BreakEvenAfter: "TP1 or immediate favorable continuation",
				TrailAfter:     "TP2 or once the extension is underway",
				EarlyExitIf: []string{
					"Delta divergence appears",
					"Momentum fades",
				},
				ForceFlatIf: []string{
					"Price accepts back through deviation and trend weakens",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Strong trend",
				"Pullback to first VWAP deviation",
			},
			RequiredConfirmations: []string{
				"Continuation confirmation",
			},
			RejectIf: []string{
				"Stop is unclear",
				"Reward-to-risk is below minimum",
			},
		},
	}
}
