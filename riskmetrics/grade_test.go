package riskmetrics

import "testing"

func TestGrade(t *testing.T) {
	cases := []struct {
		name       string
		sharpe     float64
		expectancy float64
		drawdown   float64
		want       RiskGrade
	}{
		{name: "A", sharpe: 1.1, expectancy: 0.1, drawdown: 9, want: GradeA},
		{name: "B", sharpe: 0.6, expectancy: 0.1, drawdown: 20, want: GradeB},
		{name: "C", sharpe: 0.1, expectancy: 0, drawdown: 5, want: GradeC},
		{name: "D", sharpe: 0.1, expectancy: -0.1, drawdown: 5, want: GradeD},
		{name: "F", sharpe: 0.1, expectancy: -0.1, drawdown: 21, want: GradeF},
	}

	for _, tc := range cases {
		if got := Grade(tc.sharpe, tc.expectancy, tc.drawdown); got != tc.want {
			t.Fatalf("%s grade = %s, want %s", tc.name, got, tc.want)
		}
	}
}
