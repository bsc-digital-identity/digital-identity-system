package utilities

import (
	"log"
	"os"
	"strings"
)

const (
	LanHostEnvKey = "LAN_HOST_IP"
)

// ResolveLanHost returns the LAN host used across services.
// It requires LAN_HOST_IP to be set; the application will exit if missing.
func ResolveLanHost() string {
	if host := strings.TrimSpace(os.Getenv(LanHostEnvKey)); host != "" {
		return host
	}
	log.Fatalf("environment variable %s is required to determine LAN host", LanHostEnvKey)
	return ""
}
