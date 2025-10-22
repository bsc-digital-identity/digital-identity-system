package main

import (
	"net/url"
)

func parseURL(raw string) (*url.URL, error) { return url.Parse(raw) }
func escapeQuery(s string) string           { return url.QueryEscape(s) }
