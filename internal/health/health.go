package health

import (
	"os"
	"time"
)

// GetVersion returns the application version for health checks.
// It prefers APP_VERSION from the environment and falls back to a timestamp
// so that health endpoints always return a non-empty value.
func GetVersion() string {
	if ver, ok := os.LookupEnv("APP_VERSION"); ok && ver != "" {
		return ver
	}
	return time.Now().Format(time.RFC3339Nano)
}
