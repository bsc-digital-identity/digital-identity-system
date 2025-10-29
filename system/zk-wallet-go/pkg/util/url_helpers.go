package util

import (
	"net/url"
)

func ParseURL(raw string) (*url.URL, error) { return url.Parse(raw) }
func EscapeQuery(s string) string           { return url.QueryEscape(s) }
