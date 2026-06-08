package paper

import "testing"

func TestEvaluateExitEnforcesTrailingStopsLongAndShort(t *testing.T) {
	longExit := EvaluateExit(PaperPosition{Side: "LONG", Stop: 90, TrailingStop: 101}, 100.5, 100.5, false, false, false)
	if longExit.Reason != "trailing_stop" || longExit.Action != "close" {
		t.Fatalf("expected long trailing stop close, got %+v", longExit)
	}

	shortExit := EvaluateExit(PaperPosition{Side: "SHORT", Stop: 110, TrailingStop: 99}, 99.5, 99.5, false, false, false)
	if shortExit.Reason != "trailing_stop" || shortExit.Action != "close" {
		t.Fatalf("expected short trailing stop close, got %+v", shortExit)
	}
}
