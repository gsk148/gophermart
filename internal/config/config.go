package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	RunAddress           = "RUN_ADDRESS"
	DatabaseAddress      = "DATABASE_URI"
	AccrualSystemAddress = "ACCRUAL_SYSTEM_ADDRESS"
)

type AppConfig struct {
	RunAddress           string
	DatabaseAddress      string
	AccrualSystemAddress string
}

type Config struct {
	ConfigDatabase `yaml:"database"`
}

type ConfigDatabase struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Sslmode  string `yaml:"sslmode"`
}

func MustLoad() *AppConfig {
	var err error
	config := &AppConfig{}
	getArgs(config)
	log.Print("getArgs config", config)
	getENVs(config)
	log.Print("getEnvs config", config)
	if config.DatabaseAddress == "" {
		log.Print("db address is empty, going to get default")
		config.DatabaseAddress, err = returnDefaultDB()
		log.Print("got default db address", config.DatabaseAddress)
		if err != nil {
			log.Fatal("Failed to load default DB connection")
		}
	}
	return config
}

func getArgs(cfg *AppConfig) {
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Application run address")
	flag.StringVar(&cfg.DatabaseAddress, "d", "", "Database address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	flag.Parse()
}

func getENVs(cfg *AppConfig) {
	envRunAddr := strings.TrimSpace(os.Getenv(RunAddress))
	if envRunAddr != "" {
		cfg.RunAddress = envRunAddr
	}

	databaseAddr := strings.TrimSpace(os.Getenv(DatabaseAddress))
	if databaseAddr != "" {
		cfg.DatabaseAddress = databaseAddr
	}

	accrualAddr := strings.TrimSpace(os.Getenv(AccrualSystemAddress))
	if accrualAddr != "" {
		cfg.AccrualSystemAddress = accrualAddr
	}
}

func returnDefaultDB() (string, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	connection := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Name,
		cfg.Password,
		cfg.Sslmode)
	log.Print("default connection is", connection)
	return connection, nil
}
