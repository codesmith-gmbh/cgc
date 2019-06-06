package cgcos

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func MustEnvString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env var %s not defined\n", key)
	}
	return val
}

func MustEnvInt(key string) int {
	val, err := strconv.Atoi(MustEnvString(key))
	if err != nil {
		log.Fatal(err, fmt.Sprintf("env var %s undefined or not int", key))
	}
	return val
}
