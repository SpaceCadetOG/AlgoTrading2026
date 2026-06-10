package l2recorder

import appconfig "AlgoTrading2026/config"

type RecorderConfig struct {
	Symbols         []string
	Venues          []string
	IntervalSeconds int
	MaxSnapshots    int
	OutputPath      string
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		Symbols:         []string{"BTC"},
		Venues:          []string{"hyperliquid", "aster", "lighter"},
		IntervalSeconds: 5,
		MaxSnapshots:    12,
		OutputPath:      appconfig.DataPath("l2", "l2_snapshots.csv"),
	}
}

func (c RecorderConfig) normalized() RecorderConfig {
	defaults := DefaultRecorderConfig()
	if len(c.Symbols) == 0 {
		c.Symbols = defaults.Symbols
	}
	if len(c.Venues) == 0 {
		c.Venues = defaults.Venues
	}
	if c.IntervalSeconds < 0 {
		c.IntervalSeconds = defaults.IntervalSeconds
	}
	if c.MaxSnapshots <= 0 {
		c.MaxSnapshots = defaults.MaxSnapshots
	}
	if c.OutputPath == "" {
		c.OutputPath = defaults.OutputPath
	}
	return c
}
