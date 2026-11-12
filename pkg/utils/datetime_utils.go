package CustomUtils

import "time"

func ConvertToTimeOfDay(t time.Time) time.Time {
	return time.Date(
		0, 1, 1,
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		time.UTC,
	)
}
