package paper

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func LoadState(cfg Config) (EngineState, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.StatePath), 0o755); err != nil {
		return EngineState{}, err
	}
	state := NewState(cfg)
	if body, err := os.ReadFile(cfg.StatePath); err == nil && len(body) > 0 {
		if err := json.Unmarshal(body, &state); err != nil {
			return EngineState{}, err
		}
	}
	if body, err := os.ReadFile(cfg.PositionsPath); err == nil && len(body) > 0 {
		var positions []PaperPosition
		if err := json.Unmarshal(body, &positions); err != nil {
			return EngineState{}, err
		}
		state.OpenPositions = positions
		state.OpenCount = len(positions)
	}
	return state, nil
}

func SaveState(cfg Config, state EngineState) error {
	if err := os.MkdirAll(filepath.Dir(cfg.StatePath), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfg.StatePath, append(body, '\n'), 0o644); err != nil {
		return err
	}
	positions, err := json.MarshalIndent(state.OpenPositions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfg.PositionsPath, append(positions, '\n'), 0o644)
}

func EnsureDataFiles(cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(cfg.StatePath), 0o755); err != nil {
		return err
	}
	headers := map[string][]string{
		cfg.TradesPath:  {"timestamp", "position_id", "venue", "symbol", "side", "entry_price", "exit_price", "qty", "realized_pnl", "reason"},
		cfg.EquityPath:  {"timestamp", "balance", "equity", "open_pnl", "realized_today"},
		cfg.FundingPath: {"timestamp", "position_id", "symbol", "side", "rate", "amount"},
	}
	for path, header := range headers {
		if _, err := os.Stat(path); err == nil {
			continue
		}
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		writer := csv.NewWriter(file)
		if err := writer.Write(header); err != nil {
			file.Close()
			return err
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	for _, path := range []string{cfg.EventsPath, cfg.StatePath, cfg.PositionsPath} {
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			return err
		}
	}
	return nil
}

type RuntimeSummarySnapshot struct {
	Timestamp       int64            `json:"timestamp"`
	TimeUTC         string           `json:"timeUtc"`
	Mode            string           `json:"mode"`
	ExecutionMode   string           `json:"executionMode"`
	LiveEnabled     bool             `json:"liveEnabled"`
	Summary         RuntimeSummary   `json:"summary"`
	OpenPositions   []PaperPosition  `json:"openPositions"`
	RecentClosed    []PaperPosition  `json:"recentClosed"`
	RecentDecisions []TelemetryEvent `json:"recentDecisions"`
}

func WriteRuntimeSummaryJSON(cfg Config, summary RuntimeSummary, state EngineState) error {
	now := time.Now().UTC()
	snapshot := RuntimeSummarySnapshot{
		Timestamp:       now.UnixMilli(),
		TimeUTC:         now.Format(time.RFC3339),
		Mode:            summary.Mode,
		ExecutionMode:   summary.ExecutionMode,
		LiveEnabled:     summary.LiveEnabled,
		Summary:         summary,
		OpenPositions:   append([]PaperPosition(nil), state.OpenPositions...),
		RecentClosed:    append([]PaperPosition(nil), state.RecentClosed...),
		RecentDecisions: append([]TelemetryEvent(nil), state.RecentDecisions...),
	}
	return writeJSONFile(cfg.SummaryJSONPath, snapshot)
}

func WriteRejectSummaryJSON(cfg Config, reasons map[string]int) error {
	if reasons == nil {
		reasons = map[string]int{}
	}
	return writeJSONFile(cfg.RejectSummaryJSONPath, reasons)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func AppendTrade(cfg Config, timestamp int64, position PaperPosition, exitPrice float64, qty float64, pnl float64, reason string) error {
	return appendCSVRow(cfg.TradesPath, []string{
		strconv.FormatInt(timestamp, 10),
		position.ID,
		position.Venue,
		position.Symbol,
		position.Side,
		floatString(position.EntryPrice),
		floatString(exitPrice),
		floatString(qty),
		floatString(pnl),
		reason,
	})
}

func AppendEquity(cfg Config, timestamp int64, state EngineState) error {
	return appendCSVRow(cfg.EquityPath, []string{
		strconv.FormatInt(timestamp, 10),
		floatString(state.Balance),
		floatString(state.Equity),
		floatString(state.OpenPnL),
		floatString(state.RealizedToday),
	})
}

func AppendFunding(cfg Config, event FundingEvent) error {
	return appendCSVRow(cfg.FundingPath, []string{
		strconv.FormatInt(event.Timestamp, 10),
		event.PositionID,
		event.Symbol,
		event.Side,
		floatString(event.Rate),
		floatString(event.Amount),
	})
}

func AppendEvent(cfg Config, event TelemetryEvent) error {
	file, err := os.OpenFile(cfg.EventsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = file.Write(append(body, '\n'))
	return err
}

func appendCSVRow(path string, row []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write(row); err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
