package main

import "time"

func nowUTC() TimeUTC {
	return TimeUTC{T: time.Now().UTC().Unix()}
}

func (t TimeUTC) After(other TimeUTC) bool { return t.T > other.T }
func (t TimeUTC) AddSeconds(sec int64) TimeUTC {
	return TimeUTC{T: t.T + sec}
}
