package config

import (
	"os"
)

func EthNode() string {
	return getEnv("ETHNODE", "http://localhost:8000/")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
