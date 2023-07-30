package utils

import "time"

var TimeUnixEpochStart = time.Date(
	1970,         // Year
	time.January, // Month
	0,            // Day
	0,            // Hour
	0,            // Minute
	0,            // Second
	0,            // Nanosecond
	time.UTC,     // Location
)
