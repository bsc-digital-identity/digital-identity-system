package timeutil

import (
	"time"
)

// TimeUTC is a small helper type representing Unix time (in seconds) in UTC.
// Using a dedicated type prevents confusion between local and UTC timestamps.
type TimeUTC struct{ T int64 }

func NowUTC() TimeUTC {
	return TimeUTC{T: time.Now().UTC().Unix()}
}

func (t TimeUTC) After(other TimeUTC) bool { return t.T > other.T }
func (t TimeUTC) AddSeconds(sec int64) TimeUTC {
	return TimeUTC{T: t.T + sec}
}
