package riskmetrics

type RiskGrade string

const (
	GradeA RiskGrade = "A"
	GradeB RiskGrade = "B"
	GradeC RiskGrade = "C"
	GradeD RiskGrade = "D"
	GradeF RiskGrade = "F"
)

func Grade(sharpe float64, expectancy float64, drawdownPct float64) RiskGrade {
	if expectancy < 0 && drawdownPct > 20 {
		return GradeF
	}
	if expectancy < 0 {
		return GradeD
	}
	if sharpe > 1 && expectancy > 0 && drawdownPct < 10 {
		return GradeA
	}
	if sharpe > 0.5 && expectancy > 0 {
		return GradeB
	}
	return GradeC
}
