package config

import (
	"flag"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	RunAddress           string
	DataBaseURI          string
	AccrualSystemAddress string
	TokenSecretKey       string
	TokenExpiration      time.Duration
}

func Load(envFile string) (Config, error) {
	var config = Config{}

	if len(envFile) != 0 {
		if err := godotenv.Load(envFile); err != nil {
			return config, err
		}
	}

	runAddress := flag.String("a", "", "Host Port")
	databaseURI := flag.String("d", "", "Database URI")
	accrualSystemAddress := flag.String("r", "", "Accrual System Address")
	secretKey := flag.String("s", "", "Token Secret Key")
	tokenExpiration := flag.Int("e", 24, "Token Expiration")

	flag.Parse()

	config.RunAddress = getEnvString(*runAddress, "RUN_ADDRESS")
	config.DataBaseURI = getEnvString(*databaseURI, "DATABASE_URI")
	config.AccrualSystemAddress = getEnvString(*accrualSystemAddress, "ACCRUAL_SYSTEM_ADDRESS")
	config.TokenSecretKey = getEnvString(*secretKey, "TOKEN_SECRET_KEY")

	tokenExpirationTime := getEnvInt(*tokenExpiration, "TOKEN_EXPIRATION")

	config.TokenExpiration = time.Hour * time.Duration(tokenExpirationTime)

	return config, nil
}

func getEnvString(flagValue string, envKey string) string {
	envValue, exists := os.LookupEnv(envKey)
	if exists {
		return envValue
	}
	return flagValue
}

func getEnvInt(flagValue int, envKey string) int {
	envValue, exists := os.LookupEnv(envKey)
	if exists {
		intVal, err := strconv.Atoi(envValue)
		if err != nil {
			log.Printf("cant convert env-key: %s to int", envValue)
			return 1
		}

		return intVal
	}

	return flagValue
}
