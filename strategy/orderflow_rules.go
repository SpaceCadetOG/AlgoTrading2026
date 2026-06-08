package strategy

func orderFlowRulePack() []ExecutablePlaybook {
	risk := DefaultBookRiskRule()
	return []ExecutablePlaybook{
		{
			Name:          "Unfinished Business Revisit",
			DirectionBias: "both",
			Setup:         "An unfinished high or low exists and price later revisits that level, creating a chance to trade the revisit after confirmation rather than the first print.",
			Entry: EntryRule{
				Name:       "UB Revisit Confirmation",
				SourceBook: "order_flow",
				EntryType:  "revisit_confirmation",
				Trigger:    "Enter only after the unfinished level is revisited and held or rejected with clear response.",
				ConfirmationRequired: []string{
					"Revisit of unfinished high or low",
					"Reaction candle or supporting order-flow response",
				},
				InvalidIf: []string{
					"Revisit chops through the level without clear response",
					"Liquidity disappears and the revisit loses structure",
				},
				OffsetBps: 5,
			},
			Stop: StopRule{
				Name:                   "Unfinished Level Stop",
				SourceBook:             "order_flow",
				StopType:               "order_flow_level",
				Placement:              "Beyond the unfinished level with a small buffer beyond the auction extreme.",
				BufferBps:              6,
				CatastrophicMultiplier: 1.25,
			},
			Target: TargetRule{
				Name:        "Auction Completion Targets",
				SourceBook:  "order_flow",
				TargetType:  "auction_completion",
				Target1:     "Nearest HVN or POC",
				Target2:     "Auction completion objective",
				Target3:     "VWAP or next meaningful profile barrier",
				MinRR:       risk.MinRewardToRisk,
				PartialPct1: 0.40,
				PartialPct2: 0.35,
				PartialPct3: 0.25,
			},
			Management: TradeManagementRule{
				Name:           "UB Revisit Management",
				SourceBook:     "order_flow",
				BreakEvenAfter: "First meaningful reaction away from the unfinished level",
				TrailAfter:     "Only after the auction completion move is underway",
				EarlyExitIf: []string{
					"Revisit fails and price accepts through the unfinished level",
					"Liquidity disappears before target one",
				},
				ForceFlatIf: []string{
					"Opposite unfinished-business magnet appears nearby",
				},
			},
			Risk: risk,
			RequiredContext: []string{
				"Recorded unfinished high or low",
				"Nearby footprint structure",
			},
			RequiredConfirmations: []string{
				"Revisit reaction",
			},
			RejectIf: []string{
				"No plan for stop behind the unfinished level",
				"Reward-to-risk is below minimum",
			},
		},
	}
}
