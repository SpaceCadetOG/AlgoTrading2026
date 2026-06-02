package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type VolumeProfileFinalBookPacket struct {
	Status              string   `json:"status"`
	ImplementedEngines  []string `json:"implementedEngines"`
	ImplementedSetups   []string `json:"implementedSetupStudies"`
	FeatureLabelLayer   []string `json:"featureLabelFoundation"`
	CompletedDocs       []string `json:"completedDocs"`
	InstrumentScanner   []string `json:"instrumentScanner"`
	ArchivedAudits      int      `json:"archivedAudits"`
	ArchivedFiles       []string `json:"archivedFiles"`
	NotCompleteAs       []string `json:"notCompleteAs"`
	FinalRecommendation []string `json:"finalRecommendation"`
}

func BuildVolumeProfileFinalBookPacket(archived []string) VolumeProfileFinalBookPacket {
	return VolumeProfileFinalBookPacket{
		Status: "VOLUME PROFILE BOOK COMPLETE AS RESEARCH/DOCUMENTATION LAYER",
		ImplementedEngines: []string{
			"Price Action research",
			"VWAP research",
			"Normalized L2 order book research",
			"Historical L2 snapshot recorder",
			"Volume Profile foundation",
			"Scoped and flexible profiles",
			"Profile acceptance and setup quality studies",
		},
		ImplementedSetups: []string{
			"Volume Setup #1 Accumulation",
			"Volume Setup #2 Trend",
			"Volume Setup #3 Rejection",
			"Reversal Trade Study",
			"Cross-setup comparison",
		},
		FeatureLabelLayer: []string{
			"research/setup_features_labels.csv",
			"research/setup_labels_summary.json",
			"Normalized bps labels",
			"ATR-multiple labels",
			"Acceptance labels",
			"Directional labels",
			"Triple-barrier labels",
		},
		CompletedDocs: []string{
			"docs/volume_profile_style_playbook.md",
			"docs/volume_profile_instrument_selection.md",
			"docs/volume_profile_macro_news_playbook.md",
			"docs/volume_profile_market_analysis_process.md",
			"docs/volume_profile_position_management.md",
			"docs/volume_profile_money_management.md",
			"docs/volume_profile_psychology.md",
			"docs/volume_profile_backtesting_phases.md",
			"docs/volume_profile_trading_journal.md",
			"docs/volume_profile_common_mistakes.md",
			"docs/volume_profile_real_trade_mapping.md",
			"docs/volume_profile_putting_it_together.md",
		},
		InstrumentScanner: []string{
			"research/instrument_universe.csv",
			"research/instrument_universe_summary.json",
			"Research-only scanner that ranks instruments for future research.",
			"Not a trade scanner, alert scanner, or execution scanner.",
		},
		ArchivedAudits: archivedCount(archived),
		ArchivedFiles:  archived,
		NotCompleteAs: []string{
			"live trading",
			"paper trading",
			"production execution",
			"ML system",
		},
		FinalRecommendation: []string{
			"Manual review of strongest setup filters should come next.",
			"Reversal is the first candidate for a strategy research packet.",
			"Accumulation remains important for delayed follow-through research.",
			"Historical L2 replay and tick-accurate volume-at-price remain future data improvements.",
		},
	}
}

func WriteVolumeProfileFinalBookPacketJSON(path string, packet VolumeProfileFinalBookPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeProfileFinalBookPacketMarkdown(path string, packet VolumeProfileFinalBookPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Profile Final Book Packet\n\n")
	body.WriteString(fmt.Sprintf("status=%s\n\n", packet.Status))
	writeBookPacketList(&body, "Implemented Engines", packet.ImplementedEngines)
	writeBookPacketList(&body, "Implemented Setup Studies", packet.ImplementedSetups)
	writeBookPacketList(&body, "Feature/Label Foundation", packet.FeatureLabelLayer)
	writeBookPacketList(&body, "Completed Docs", packet.CompletedDocs)
	writeBookPacketList(&body, "Instrument Scanner", packet.InstrumentScanner)
	body.WriteString(fmt.Sprintf("## Files Archived\n\narchivedAudits=%d\n\n", packet.ArchivedAudits))
	for _, file := range packet.ArchivedFiles {
		body.WriteString("- " + file + "\n")
	}
	body.WriteString("\n")
	writeBookPacketList(&body, "Not Complete As", packet.NotCompleteAs)
	writeBookPacketList(&body, "Final Recommendation", packet.FinalRecommendation)
	return os.WriteFile(path, []byte(body.String()), 0644)
}

func archivedCount(archived []string) int {
	count := 0
	for _, file := range archived {
		if file != "" {
			count++
		}
	}
	return count
}
