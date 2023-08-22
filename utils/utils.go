package utils

import (
	"os"
	"strconv"
	"time"
)

func GetIntEnv(key string, defaultValue ...int) int {
	valStr := os.Getenv(key)
	val, err := strconv.Atoi(valStr)
	if err != nil {
		if defaultValue != nil {
			return defaultValue[0]
		}
		return 0
	}

	return val
}

func GetCacheDurationEnv() time.Duration {
	seconds := GetIntEnv("CACHE_DURATION_SECONDS", 1)
	return time.Duration(seconds)
}
