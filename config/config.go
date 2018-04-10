package config

import (
	"os"
	"strconv"
)

func EthNode() string {
	return getEnv("ETHNODE", "http://localhost:8000/")
}

func BtcNode() string {
	return getEnv("BTCNODE", "http://localhost:8332/")
}

func BtcRpcUser() string {
	return getEnv("BTCRPCUSER", "bitcoin")
}

func BtcRpcPass() string {
	return getEnv("BTCRPCPASS", "bitcoin")
}

func CommitInterval() uint64 {
	returnVal, _ := strconv.ParseUint(getEnv("COMMIT_INTERVAL", "240"), 10, 64)
	return returnVal
}

func CoinType() byte {
	coinType := getEnv("COINTYPE", "regtest")

	switch coinType {
	case "regtest":
		return 0x6F
	default:
		return 0x00
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
