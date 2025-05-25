package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var AppConfig *Config

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`

	Encryption struct {
		Secret  string `yaml:"secret"`
		HMACKey string `yaml:"hmac_key"`
	} `yaml:"encryption"`

	Database struct {
		Host           string `yaml:"host"`
		Port           int    `yaml:"port"`
		User           string `yaml:"user"`
		Password       string `yaml:"password"`
		DBName         string `yaml:"dbname"`
		SSLMode        string `yaml:"sslmode"`
		MigrationsPath string `yaml:"migrations_path"`
	} `yaml:"database"`
}

func Load(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	AppConfig = &cfg
	return nil
}
