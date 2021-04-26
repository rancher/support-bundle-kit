package utils

import (
	"os"
	"strconv"
	"time"
)

func EnvGetBool(key string, defaultValue bool) bool {
	if parsed, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return parsed
	}
	return defaultValue
}

func EnvGetInt(key string, defaultValue int) int {
	if parsed, err := strconv.ParseInt(os.Getenv(key), 10, 32); err == nil {
		return int(parsed)
	}
	return defaultValue
}

func EnvGetDuration(key string, defaultValue time.Duration) time.Duration {
	if parsed, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return parsed
	}
	return defaultValue
}
