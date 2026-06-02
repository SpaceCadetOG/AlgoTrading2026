package research

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strconv"
)

type ProfileAcceptanceQualityRow struct {
	Category           string  `json:"category"`
	Name               string  `json:"name"`
	TotalProfiles      int     `json:"totalProfiles"`
	AcceptedProfiles   int     `json:"acceptedProfiles"`
	RejectedProfiles   int     `json:"rejectedProfiles"`
	NeutralProfiles    int     `json:"neutralProfiles"`
	AcceptanceRate     float64 `json:"acceptanceRate"`
	DominantShape      string  `json:"dominantShape"`
	AverageConfidence  float64 `json:"averageConfidence"`
	AverageHVNCount    float64 `json:"averageHvnCount"`
	AverageLVNCount    float64 `json:"averageLvnCount"`
	AverageVolume      float64 `json:"averageVolume"`
	AveragePOCDistance float64 `json:"averagePocDistance"`
}

type ProfileAcceptanceMetricSummary struct {
	Profiles           int     `json:"profiles"`
	DominantShape      string  `json:"dominantShape"`
	AverageConfidence  float64 `json:"averageConfidence"`
	AverageHVNCount    float64 `json:"averageHvnCount"`
	AverageLVNCount    float64 `json:"averageLvnCount"`
	AverageVolume      float64 `json:"averageVolume"`
	AveragePOCDistance float64 `json:"averagePocDistance"`
}

type ProfileAcceptanceQualitySummary struct {
	AcceptedProfiles          int                            `json:"acceptedProfiles"`
	RejectedProfiles          int                            `json:"rejectedProfiles"`
	AcceptedMetrics           ProfileAcceptanceMetricSummary `json:"acceptedMetrics"`
	RejectedMetrics           ProfileAcceptanceMetricSummary `json:"rejectedMetrics"`
	BestProfileType           string                         `json:"bestProfileType"`
	BestProfileAcceptanceRate float64                        `json:"bestProfileAcceptanceRate"`
	BestShape                 string                         `json:"bestShape"`
	BestShapeAcceptanceRate   float64                        `json:"bestShapeAcceptanceRate"`
	BestScope                 string                         `json:"bestScope"`
	BestScopeAcceptanceRate   float64                        `json:"bestScopeAcceptanceRate"`
	Rows                      []ProfileAcceptanceQualityRow  `json:"rows"`
	Notes                     []string                       `json:"notes"`
}

func BuildProfileAcceptanceQualityFromFiles(flexiblePath string, shapePath string, scopedPath string) ([]ProfileAcceptanceQualityRow, ProfileAcceptanceQualitySummary, error) {
	flexible, err := readFlexibleProfileRecords(flexiblePath)
	if err != nil {
		return nil, ProfileAcceptanceQualitySummary{}, err
	}
	scopeWindows, err := readScopedProfileWindows(scopedPath)
	if err != nil {
		return nil, ProfileAcceptanceQualitySummary{}, err
	}
	if _, err := readShapeStudyScopes(shapePath); err != nil {
		return nil, ProfileAcceptanceQualitySummary{}, err
	}
	rows := BuildProfileAcceptanceQualityRows(flexible, scopeWindows)
	summary := BuildProfileAcceptanceQualitySummary(rows)
	return rows, summary, nil
}

func BuildProfileAcceptanceQualityRows(flexible []flexibleProfileRecord, scopes []profileScopeWindow) []ProfileAcceptanceQualityRow {
	groups := map[string]*acceptanceAccumulator{}
	accepted := newAcceptanceAccumulator("acceptance_state", "accepted")
	rejected := newAcceptanceAccumulator("acceptance_state", "rejected")

	for _, row := range flexible {
		addAcceptanceGroup(groups, "profile_type", row.ProfileType).add(row)
		addAcceptanceGroup(groups, "profile_shape", row.ProfileShape).add(row)
		for _, scope := range scopesForRecord(row, scopes) {
			addAcceptanceGroup(groups, "profile_scope", scope).add(row)
		}
		switch row.AcceptanceState {
		case "accepted":
			accepted.add(row)
		case "rejected":
			rejected.add(row)
		}
	}

	out := make([]ProfileAcceptanceQualityRow, 0, len(groups)+2)
	for _, group := range groups {
		out = append(out, group.row())
	}
	out = append(out, accepted.row(), rejected.row())
	sort.Slice(out, func(i int, j int) bool {
		if out[i].Category == out[j].Category {
			return out[i].Name < out[j].Name
		}
		return out[i].Category < out[j].Category
	})
	return out
}

func BuildProfileAcceptanceQualitySummary(rows []ProfileAcceptanceQualityRow) ProfileAcceptanceQualitySummary {
	summary := ProfileAcceptanceQualitySummary{
		Rows: rows,
		Notes: []string{
			"Acceptance quality is calculated from flexible event-anchored profiles.",
			"Average POC distance uses absolute distance between POC and value-area midpoint because the profile artifact does not store an execution price.",
			"Profile scope acceptance assigns flexible profiles to scoped windows when the flexible profile is fully contained inside the scope window.",
		},
	}
	for _, row := range rows {
		if row.Category == "acceptance_state" && row.Name == "accepted" {
			summary.AcceptedProfiles = row.TotalProfiles
			summary.AcceptedMetrics = metricSummaryFromRow(row)
		}
		if row.Category == "acceptance_state" && row.Name == "rejected" {
			summary.RejectedProfiles = row.TotalProfiles
			summary.RejectedMetrics = metricSummaryFromRow(row)
		}
		if row.Category == "profile_type" && row.TotalProfiles > 0 && row.AcceptanceRate > summary.BestProfileAcceptanceRate {
			summary.BestProfileType = row.Name
			summary.BestProfileAcceptanceRate = row.AcceptanceRate
		}
		if row.Category == "profile_shape" && row.TotalProfiles > 0 && row.AcceptanceRate > summary.BestShapeAcceptanceRate {
			summary.BestShape = row.Name
			summary.BestShapeAcceptanceRate = row.AcceptanceRate
		}
		if row.Category == "profile_scope" && row.TotalProfiles > 0 && row.AcceptanceRate > summary.BestScopeAcceptanceRate {
			summary.BestScope = row.Name
			summary.BestScopeAcceptanceRate = row.AcceptanceRate
		}
	}
	return summary
}

func WriteProfileAcceptanceQualityCSV(path string, rows []ProfileAcceptanceQualityRow) error {
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
		"category",
		"name",
		"total_profiles",
		"accepted_profiles",
		"rejected_profiles",
		"neutral_profiles",
		"acceptance_rate",
		"dominant_shape",
		"average_confidence",
		"average_hvn_count",
		"average_lvn_count",
		"average_volume",
		"average_distance_from_poc",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Category,
			row.Name,
			strconv.Itoa(row.TotalProfiles),
			strconv.Itoa(row.AcceptedProfiles),
			strconv.Itoa(row.RejectedProfiles),
			strconv.Itoa(row.NeutralProfiles),
			floatToString(row.AcceptanceRate),
			row.DominantShape,
			floatToString(row.AverageConfidence),
			floatToString(row.AverageHVNCount),
			floatToString(row.AverageLVNCount),
			floatToString(row.AverageVolume),
			floatToString(row.AveragePOCDistance),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteProfileAcceptanceQualitySummaryJSON(path string, summary ProfileAcceptanceQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

type flexibleProfileRecord struct {
	ProfileType     string
	ProfileID       string
	StartTime       int64
	EndTime         int64
	Symbol          string
	POC             float64
	VAH             float64
	VAL             float64
	HVNCount        int
	LVNCount        int
	ProfileShape    string
	ShapeConfidence float64
	AcceptanceState string
	ProfileVolume   float64
}

type profileScopeWindow struct {
	Scope        string
	ProfileStart int64
	ProfileEnd   int64
}

type acceptanceAccumulator struct {
	category      string
	name          string
	total         int
	accepted      int
	rejected      int
	neutral       int
	confidenceSum float64
	hvnSum        float64
	lvnSum        float64
	volumeSum     float64
	pocDistSum    float64
	shapeCounts   map[string]int
}

func newAcceptanceAccumulator(category string, name string) *acceptanceAccumulator {
	return &acceptanceAccumulator{category: category, name: name, shapeCounts: map[string]int{}}
}

func addAcceptanceGroup(groups map[string]*acceptanceAccumulator, category string, name string) *acceptanceAccumulator {
	key := category + ":" + name
	if groups[key] == nil {
		groups[key] = newAcceptanceAccumulator(category, name)
	}
	return groups[key]
}

func (a *acceptanceAccumulator) add(row flexibleProfileRecord) {
	a.total++
	switch row.AcceptanceState {
	case "accepted":
		a.accepted++
	case "rejected":
		a.rejected++
	default:
		a.neutral++
	}
	a.confidenceSum += row.ShapeConfidence
	a.hvnSum += float64(row.HVNCount)
	a.lvnSum += float64(row.LVNCount)
	a.volumeSum += row.ProfileVolume
	a.pocDistSum += absResearch(row.POC - ((row.VAH + row.VAL) / 2))
	a.shapeCounts[row.ProfileShape]++
}

func (a *acceptanceAccumulator) row() ProfileAcceptanceQualityRow {
	if a.total == 0 {
		return ProfileAcceptanceQualityRow{Category: a.category, Name: a.name}
	}
	return ProfileAcceptanceQualityRow{
		Category:           a.category,
		Name:               a.name,
		TotalProfiles:      a.total,
		AcceptedProfiles:   a.accepted,
		RejectedProfiles:   a.rejected,
		NeutralProfiles:    a.neutral,
		AcceptanceRate:     float64(a.accepted) / float64(a.total),
		DominantShape:      dominantShape(a.shapeCounts),
		AverageConfidence:  a.confidenceSum / float64(a.total),
		AverageHVNCount:    a.hvnSum / float64(a.total),
		AverageLVNCount:    a.lvnSum / float64(a.total),
		AverageVolume:      a.volumeSum / float64(a.total),
		AveragePOCDistance: a.pocDistSum / float64(a.total),
	}
}

func metricSummaryFromRow(row ProfileAcceptanceQualityRow) ProfileAcceptanceMetricSummary {
	return ProfileAcceptanceMetricSummary{
		Profiles:           row.TotalProfiles,
		DominantShape:      row.DominantShape,
		AverageConfidence:  row.AverageConfidence,
		AverageHVNCount:    row.AverageHVNCount,
		AverageLVNCount:    row.AverageLVNCount,
		AverageVolume:      row.AverageVolume,
		AveragePOCDistance: row.AveragePOCDistance,
	}
}

func dominantShape(counts map[string]int) string {
	best := ""
	bestCount := -1
	for shape, count := range counts {
		if count > bestCount || count == bestCount && shape < best {
			best = shape
			bestCount = count
		}
	}
	return best
}

func scopesForRecord(row flexibleProfileRecord, scopes []profileScopeWindow) []string {
	out := make([]string, 0)
	for _, scope := range scopes {
		if row.StartTime >= scope.ProfileStart && row.EndTime <= scope.ProfileEnd {
			out = append(out, scope.Scope)
		}
	}
	return out
}

func readFlexibleProfileRecords(path string) ([]flexibleProfileRecord, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]flexibleProfileRecord, 0, len(records))
	for _, record := range records {
		out = append(out, flexibleProfileRecord{
			ProfileType:     record["profile_type"],
			ProfileID:       record["profile_id"],
			StartTime:       parseIntRecord(record["start_time"]),
			EndTime:         parseIntRecord(record["end_time"]),
			Symbol:          record["symbol"],
			POC:             parseFloatRecord(record["poc"]),
			VAH:             parseFloatRecord(record["vah"]),
			VAL:             parseFloatRecord(record["val"]),
			HVNCount:        int(parseIntRecord(record["hvn_count"])),
			LVNCount:        int(parseIntRecord(record["lvn_count"])),
			ProfileShape:    record["profile_shape"],
			ShapeConfidence: parseFloatRecord(record["shape_confidence"]),
			AcceptanceState: record["acceptance_state"],
			ProfileVolume:   parseFloatRecord(record["profile_volume"]),
		})
	}
	return out, nil
}

func readScopedProfileWindows(path string) ([]profileScopeWindow, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]profileScopeWindow, 0)
	for _, record := range records {
		scope := record["scope"]
		key := scope + ":" + record["profile_start"] + ":" + record["profile_end"]
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, profileScopeWindow{
			Scope:        scope,
			ProfileStart: parseIntRecord(record["profile_start"]),
			ProfileEnd:   parseIntRecord(record["profile_end"]),
		})
	}
	return out, nil
}

func readShapeStudyScopes(path string) (map[string]string, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, record := range records {
		out[record["scope"]] = record["profile_shape"]
	}
	return out, nil
}

func readCSVRecords(path string) ([]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("csv has no header")
	}
	header := rows[0]
	out := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		record := map[string]string{}
		for i, key := range header {
			if i < len(row) {
				record[key] = row[i]
			}
		}
		out = append(out, record)
	}
	return out, nil
}

func parseFloatRecord(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func parseIntRecord(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func absResearch(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
