package webutil

import (
	"log/slog"
	"time"
)

func ToTimeZone(t time.Time, name string) time.Time {
	loc, err := time.LoadLocation(name)
	if err != nil {
		slog.Error("cannot load location", "name", name, "err", err)
		return t
	}
	return t.In(loc)
}
