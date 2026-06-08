package tradetape

import (
	"fmt"
	"sort"
	"time"
)

type Fetcher func() ([]TradeTapePrint, error)

type VenueFetcher struct {
	Venue  string
	Symbol string
	Fetch  Fetcher
}

type Recorder struct {
	Config   RecorderConfig
	Fetchers []VenueFetcher
}

type RecorderSummary struct {
	Rounds int
	Rows   int
	Errors int
}

func NewRecorder(config RecorderConfig, fetchers []VenueFetcher) *Recorder {
	return &Recorder{
		Config:   config,
		Fetchers: append([]VenueFetcher(nil), fetchers...),
	}
}

func (r *Recorder) RecordRound() (RecorderSummary, error) {
	rows := make([]TradeTapeRow, 0)
	fetchers := append([]VenueFetcher(nil), r.Fetchers...)
	sort.Slice(fetchers, func(i int, j int) bool {
		if fetchers[i].Venue == fetchers[j].Venue {
			return fetchers[i].Symbol < fetchers[j].Symbol
		}
		return fetchers[i].Venue < fetchers[j].Venue
	})
	for _, fetcher := range fetchers {
		if fetcher.Fetch == nil {
			rows = append(rows, ErrorRow(fetcher.Venue, fetcher.Symbol, fmt.Errorf("missing trade tape fetcher")))
			continue
		}
		prints, err := fetcher.Fetch()
		if err != nil {
			rows = append(rows, ErrorRow(fetcher.Venue, fetcher.Symbol, err))
			continue
		}
		for _, print := range prints {
			rows = append(rows, RowFromPrint(print))
		}
	}
	if err := AppendRows(r.Config.OutputPath, rows); err != nil {
		return RecorderSummary{}, err
	}
	summary := RecorderSummary{Rounds: 1, Rows: len(rows)}
	for _, row := range rows {
		if !row.Valid || row.Error != "" {
			summary.Errors++
		}
	}
	return summary, nil
}

func (r *Recorder) Run() (RecorderSummary, error) {
	maxRounds := r.Config.MaxRounds
	if maxRounds <= 0 {
		maxRounds = 1
	}
	interval := r.Config.IntervalSeconds
	if interval <= 0 {
		interval = 5
	}
	var total RecorderSummary
	for round := 0; round < maxRounds; round++ {
		summary, err := r.RecordRound()
		if err != nil {
			return total, err
		}
		total.Rounds += summary.Rounds
		total.Rows += summary.Rows
		total.Errors += summary.Errors
		if round < maxRounds-1 {
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
	return total, nil
}
