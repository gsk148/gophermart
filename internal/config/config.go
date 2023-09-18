package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ConfigDatabase `yaml:"database"`
}

type ConfigDatabase struct {
	Port     string `yaml:"port" env:"PORT" env-default:"5432"`
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Name     string `yaml:"name" env:"NAME" env-default:"postgres"`
	Username string `yaml:"username"  env-default:"default"`
	Password string `yaml:"password" env:"PASSWORD"`
	Sslmode  string `yaml:"sslmode" env-default:"disable"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadConfig("config/config.yaml", &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
