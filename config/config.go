package config

import (
	"os"
	"strconv"
)

func EthNode() string {
	return getEnv("ETHNODE", "http://localhost:8000/")
}

func CommitInterval() uint64 {
	returnVal, _ := strconv.ParseUint(getEnv("COMMIT_INTERVAL", "240"), 10, 64)
	return returnVal
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
