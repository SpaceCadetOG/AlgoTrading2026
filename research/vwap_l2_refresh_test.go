package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/orderbook"
)

func TestL2QualityClassifiers(t *testing.T) {
	if got := ClassifySpreadQuality(0.01, 0.02, 0.05); got != SpreadQualityTight {
		t.Fatalf("spread quality=%s want tight", got)
	}
	if got := ClassifySpreadQuality(0.04, 0.02, 0.05); got != SpreadQualityNormal {
		t.Fatalf("spread quality=%s want normal", got)
	}
	if got := ClassifySpreadQuality(0.06, 0.02, 0.05); got != SpreadQualityWide {
		t.Fatalf("spread quality=%s want wide", got)
	}
	if got := ClassifyBookPressure(0.2); got != BookPressureBid {
		t.Fatalf("book pressure=%s want bid", got)
	}
	if got := ClassifyBookPressure(-0.2); got != BookPressureAsk {
		t.Fatalf("book pressure=%s want ask", got)
	}
	if got := ClassifyBookPressure(0.01); got != BookPressureBalanced {
		t.Fatalf("book pressure=%s want balanced", got)
	}
	if got := ClassifyLiquidityQuality(0); got != LiquidityQualityNoneNearVWAP {
		t.Fatalf("liquidity quality=%s want none", got)
	}
	if got := ClassifyLiquidityQuality(50_000); got != LiquidityQualityWeakNearVWAP {
		t.Fatalf("liquidity quality=%s want weak", got)
	}
	if got := ClassifyLiquidityQuality(150_000); got != LiquidityQualityStrongNearVWAP {
		t.Fatalf("liquidity quality=%s want strong", got)
	}
}

func TestBuildVWAPL2RefreshRows(t *testing.T) {
	result := BuildVWAPL2Refresh([]VWAPL2RefreshInput{{
		Venue:    "hyperliquid",
		Symbol:   "BTC",
		Candles:  testVWAPL2Candles(),
		Snapshot: testVWAPL2Snapshot("hyperliquid", "BTC", "99", "10", "101", "1"),
	}})
	if len(result.FeatureRows) != 4 {
		t.Fatalf("feature rows=%d want 4", len(result.FeatureRows))
	}
	if len(result.ContextRows) != 4 {
		t.Fatalf("context rows=%d want 4", len(result.ContextRows))
	}
	row := result.FeatureRows[0]
	if row.Venue != "hyperliquid" || row.Symbol != "BTC" {
		t.Fatalf("bad metadata: %+v", row)
	}
	if row.SpreadPct == 0 || row.BidDepth1Pct == 0 || row.AskDepth1Pct == 0 {
		t.Fatalf("missing l2 metrics: %+v", row)
	}
	if !result.Summary.SnapshotOnly || len(result.Summary.Notes) == 0 {
		t.Fatalf("expected snapshot limitation note: %+v", result.Summary)
	}
}

func TestVWAPL2BehaviorSummaryClassification(t *testing.T) {
	candles := []exchanges.Candle{
		{High: "100", Low: "100", Close: "101", Volume: "1", StartTime: 1},
		{High: "103", Low: "99", Close: "101", Volume: "1", StartTime: 2},
		{High: "103", Low: "99", Close: "99", Volume: "1", StartTime: 3},
	}
	bidSupport := testVWAPL2Snapshot("hyperliquid", "BTC", "99", "10", "101", "1")
	bidCollapse := testVWAPL2Snapshot("hyperliquid", "BTC", "99", "1", "101", "10")

	bounce := BuildVWAPL2Refresh([]VWAPL2RefreshInput{{Venue: "hyperliquid", Symbol: "BTC", Candles: candles[:2], Snapshot: bidSupport}})
	if bounce.Summary.Bounces != 1 {
		t.Fatalf("bounces=%d want 1 summary=%+v", bounce.Summary.Bounces, bounce.Summary)
	}
	if bounce.Summary.BounceBidSupport != 1 || bounce.Summary.BounceAskWeakness != 1 {
		t.Fatalf("expected bounce l2 tags: %+v", bounce.Summary)
	}

	breakResult := BuildVWAPL2Refresh([]VWAPL2RefreshInput{{Venue: "hyperliquid", Symbol: "BTC", Candles: candles, Snapshot: bidCollapse}})
	if breakResult.Summary.Breaks != 1 {
		t.Fatalf("breaks=%d want 1 summary=%+v", breakResult.Summary.Breaks, breakResult.Summary)
	}
	if breakResult.Summary.BreakAskPressure != 1 || breakResult.Summary.BreakBidCollapse != 1 {
		t.Fatalf("expected break l2 tags: %+v", breakResult.Summary)
	}
}

func TestVWAPL2RefreshWriters(t *testing.T) {
	dir := t.TempDir()
	result := BuildVWAPL2Refresh([]VWAPL2RefreshInput{{
		Venue:    "aster",
		Symbol:   "BTCUSDT",
		Candles:  testVWAPL2Candles(),
		Snapshot: testVWAPL2Snapshot("aster", "BTCUSDT", "99", "10", "101", "1"),
	}})

	featuresPath := filepath.Join(dir, "vwap_features_l2.csv")
	contextPath := filepath.Join(dir, "context_features_l2.csv")
	summaryPath := filepath.Join(dir, "vwap_behavior_l2_summary.json")

	if err := WriteVWAPL2FeaturesCSV(featuresPath, result.FeatureRows); err != nil {
		t.Fatalf("write features: %v", err)
	}
	assertContains(t, featuresPath, "timestamp,venue,symbol,close,volume,session_vwap,distance_from_vwap,vwap_slope,vwap_regime,spread_pct,imbalance_1pct,bid_depth_1pct,ask_depth_1pct,liquidity_near_vwap")
	assertContains(t, featuresPath, "aster")

	if err := WriteContextL2FeaturesCSV(contextPath, result.ContextRows); err != nil {
		t.Fatalf("write context: %v", err)
	}
	assertContains(t, contextPath, "timestamp,venue,symbol,close,session_vwap,distance_from_vwap,ema9,ema20,ema_alignment,vwap_slope,trend_alignment,spread_pct,imbalance_1pct,bid_depth_1pct,ask_depth_1pct,liquidity_near_vwap,spread_quality,book_pressure,liquidity_quality")
	assertContains(t, contextPath, "bid_pressure")

	if err := WriteVWAPL2BehaviorSummaryJSON(summaryPath, result.Summary); err != nil {
		t.Fatalf("write summary: %v", err)
	}
	body, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if !strings.Contains(string(body), "snapshotOnly") {
		t.Fatalf("missing snapshot limitation field: %s", body)
	}
	var decoded VWAPL2BehaviorSummary
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}

func testVWAPL2Candles() []exchanges.Candle {
	return []exchanges.Candle{
		{High: "100", Low: "100", Close: "100", Volume: "1", StartTime: 1},
		{High: "103", Low: "99", Close: "101", Volume: "2", StartTime: 2},
		{High: "102", Low: "98", Close: "99", Volume: "2", StartTime: 3},
		{High: "101", Low: "99", Close: "100", Volume: "1", StartTime: 4},
	}
}

func testVWAPL2Snapshot(venue string, symbol string, bidPrice string, bidSize string, askPrice string, askSize string) orderbook.OrderBookSnapshot {
	return orderbook.OrderBookSnapshot{
		Venue:  venue,
		Symbol: symbol,
		Time:   10,
		Bids: []orderbook.BookLevel{
			{Price: bidPrice, Size: bidSize},
		},
		Asks: []orderbook.BookLevel{
			{Price: askPrice, Size: askSize},
		},
		IsSnapshot: true,
	}
}
