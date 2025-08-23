package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	// tenta carregar o .env se existir
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env file not found, relying on system environment variables")
	} else {
		log.Println("✅ .env loaded")
	}
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	valueAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valueAsInt
}
