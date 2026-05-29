package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/riskmetrics"
)

type StrategyRiskMetrics struct {
	Strategy string `json:"strategy"`

	Trades  int     `json:"trades"`
	WinRate float64 `json:"winRate"`

	NetPnL float64 `json:"netPnL"`

	Sharpe  float64 `json:"sharpe"`
	Sortino float64 `json:"sortino"`

	Expectancy float64 `json:"expectancy"`

	Variance float64 `json:"variance"`
	StdDev   float64 `json:"stdDev"`

	MaxDrawdown    float64 `json:"maxDrawdown"`
	MaxDrawdownPct float64 `json:"maxDrawdownPct"`

	AverageHoldCandles float64 `json:"averageHoldCandles"`
	AverageHoldHours   float64 `json:"averageHoldHours"`

	TradesPerDay   float64 `json:"tradesPerDay"`
	TradesPerWeek  float64 `json:"tradesPerWeek"`
	TradesPerMonth float64 `json:"tradesPerMonth"`

	RiskGrade riskmetrics.RiskGrade `json:"riskGrade"`
}

type Chapter6RiskSummary struct {
	TopSharpeStrategy        string                `json:"topSharpeStrategy"`
	TopSharpe                float64               `json:"topSharpe"`
	WorstRiskGradeStrategy   string                `json:"worstRiskGradeStrategy"`
	WorstRiskGrade           riskmetrics.RiskGrade `json:"worstRiskGrade"`
	LargestExecutionRate     float64               `json:"largestExecutionRate"`
	LargestExecutionStrategy string                `json:"largestExecutionStrategy"`
	Metrics                  []StrategyRiskMetrics `json:"metrics"`
}

func WriteChapter6RiskMetricsCSV(path string, rows []StrategyRiskMetrics) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		"strategy",
		"trades",
		"win_rate",
		"net_pnl",
		"sharpe",
		"sortino",
		"expectancy",
		"variance",
		"stddev",
		"max_drawdown",
		"max_drawdown_pct",
		"average_hold_candles",
		"average_hold_hours",
		"trades_per_day",
		"trades_per_week",
		"trades_per_month",
		"risk_grade",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.Strategy,
			strconv.Itoa(row.Trades),
			floatToString(row.WinRate),
			floatToString(row.NetPnL),
			floatToString(row.Sharpe),
			floatToString(row.Sortino),
			floatToString(row.Expectancy),
			floatToString(row.Variance),
			floatToString(row.StdDev),
			floatToString(row.MaxDrawdown),
			floatToString(row.MaxDrawdownPct),
			floatToString(row.AverageHoldCandles),
			floatToString(row.AverageHoldHours),
			floatToString(row.TradesPerDay),
			floatToString(row.TradesPerWeek),
			floatToString(row.TradesPerMonth),
			string(row.RiskGrade),
		}); err != nil {
			return err
		}
	}

	return writer.Error()
}

func WriteChapter6RiskSummaryJSON(path string, summary Chapter6RiskSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
