package realism

type RealismGrade string

const (
	GradeA RealismGrade = "A"
	GradeB RealismGrade = "B"
	GradeC RealismGrade = "C"
	GradeD RealismGrade = "D"
	GradeF RealismGrade = "F"
)

func GradeRealism(assumptions RealismAssumptions, severity DislocationSeverity) RealismGrade {
	missing := assumptions.MissingCount()

	if severity == SeverityHigh && missing >= 5 {
		return GradeF
	}
	if missing == 0 && severity == SeverityLow {
		return GradeA
	}
	if missing <= 2 && severity != SeverityHigh {
		return GradeB
	}
	if missing <= 4 && severity != SeverityHigh {
		return GradeC
	}
	if missing <= 7 {
		return GradeD
	}
	return GradeF
}
