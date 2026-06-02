package volumeprofile

import "testing"

func TestAnalyzeProfileShapeDProfile(t *testing.T) {
	profile := shapeStudyProfile([]float64{20, 40, 100, 40, 20}, 25)
	study := AnalyzeProfileShape(profile)
	if study.ProfileShape != ShapeDProfile || study.ShapeReason != ShapeReasonBalancedDistribution {
		t.Fatalf("study=%+v", study)
	}
	if study.ShapeConfidence <= 0 {
		t.Fatalf("expected confidence: %+v", study)
	}
}

func TestAnalyzeProfileShapePProfile(t *testing.T) {
	profile := shapeStudyProfile([]float64{10, 20, 30, 80, 100}, 25)
	study := AnalyzeProfileShape(profile)
	if study.ProfileShape != ShapePProfile || study.ShapeReason != ShapeReasonUpperDistributionDominant {
		t.Fatalf("study=%+v", study)
	}
	if study.DistributionBalance <= 0 {
		t.Fatalf("expected positive balance: %+v", study)
	}
}

func TestAnalyzeProfileShapeBProfile(t *testing.T) {
	profile := shapeStudyProfile([]float64{100, 80, 30, 20, 10}, 25)
	study := AnalyzeProfileShape(profile)
	if study.ProfileShape != ShapeBProfile || study.ShapeReason != ShapeReasonLowerDistributionDominant {
		t.Fatalf("study=%+v", study)
	}
	if study.DistributionBalance >= 0 {
		t.Fatalf("expected negative balance: %+v", study)
	}
}

func TestAnalyzeProfileShapeThinProfile(t *testing.T) {
	profile := shapeStudyProfile([]float64{10, 11, 10, 9, 10}, 25)
	study := AnalyzeProfileShape(profile)
	if study.ProfileShape != ShapeThinProfile {
		t.Fatalf("study=%+v", study)
	}
}

func shapeStudyProfile(volumes []float64, poc float64) VolumeProfile {
	bins := make([]PriceBin, 0, len(volumes))
	for i, volume := range volumes {
		low := float64(i * 10)
		bins = append(bins, PriceBin{
			Low:    low,
			High:   low + 10,
			Mid:    low + 5,
			Volume: volume,
		})
	}
	return VolumeProfile{
		Bins:        bins,
		POC:         poc,
		TotalVolume: TotalVolume(bins),
		LVNs:        LowVolumeNodes(bins),
		HVNs:        HighVolumeNodes(bins),
	}
}
