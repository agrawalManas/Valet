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

func IsSameDate(startDate, endDate time.Time) bool {
	year1, month1, day1 := startDate.Date()
	year2, month2, day2 := endDate.Date()

	return year1 == year2 && month1 == month2 && day1 == day2
}

func IsDateBefore(startDate, endDate time.Time) bool {
	y1, m1, d1 := startDate.Date()
	y2, m2, d2 := endDate.Date()

	return time.Date(y1, m1, d1, 0, 0, 0, 0, startDate.Location()).
		Before(time.Date(y2, m2, d2, 0, 0, 0, 0, endDate.Location()))
}

func IsDateAfter(startDate, endDate time.Time) bool {
	return IsDateBefore(endDate, startDate)
}
