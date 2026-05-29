package indicators

import "time"

func HourOfDay(times []int64) []int {
	out := make([]int, len(times))
	for i, timestamp := range times {
		out[i] = time.UnixMilli(timestamp).UTC().Hour()
	}
	return out
}

func DayOfWeek(times []int64) []int {
	out := make([]int, len(times))
	for i, timestamp := range times {
		out[i] = int(time.UnixMilli(timestamp).UTC().Weekday())
	}
	return out
}
