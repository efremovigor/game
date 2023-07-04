package env

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

const envPath = ".env"

func GetEnvVariable(key string) string {

	err := godotenv.Load(envPath)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
