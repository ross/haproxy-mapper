package main

import (
	"time"
)

var isoFormat = "2006-01-02T15:04:05-0700"

func EpochToIso8601(ts uint) string {
	return time.Unix(int64(ts), 0).Format(isoFormat)
}

func NowIso8610() string {
	return time.Now().Format(isoFormat)
}
