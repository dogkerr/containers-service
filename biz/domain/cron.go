package domain

type TimeFormat string

const (
	Month  TimeFormat = "MONTH"
	Day    TimeFormat = "DAY"
	Hour   TimeFormat = "HOUR"
	Minute TimeFormat = "MINUTE"
	Second TimeFormat = "SECOND"
)

var GetTimeFormat = map[string]TimeFormat{ //create a map to link enumeration values with string representation
	"MONTH":  Month,
	"DAY":    Day,
	"HOUR":   Hour,
	"MINUTE": Minute,
	"SECOND": Second,
}
