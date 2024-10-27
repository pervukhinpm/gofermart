package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DataBaseURI          string
	AccrualSystemAddress string
}

func NewConfig() *Config {

	var config = Config{}

	flag.StringVar(&config.RunAddress, "a", "localhost:8080", "Host Port")
	flag.StringVar(&config.DataBaseURI, "d", "user=postgres password=adm dbname=postgres host=localhost port=5432 sslmode=disable", "Database URI")
	flag.StringVar(&config.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual System Address")

	flag.Parse()

	if runAddressEnv := os.Getenv("RUN_ADDRESS"); runAddressEnv != "" {
		config.RunAddress = runAddressEnv
	}

	if dataBaseURIEnv := os.Getenv("DATABASE_URI"); dataBaseURIEnv != "" {
		config.DataBaseURI = dataBaseURIEnv
	}

	if accrualSystemAddressEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSystemAddressEnv != "" {
		config.AccrualSystemAddress = accrualSystemAddressEnv
	}

	return &config
}
