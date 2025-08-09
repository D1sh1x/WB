package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string     `yaml:"env"`
	Database   Database   `yaml:"database"`
	HTTPServer HTTPServer `yaml:"http_server"`
	Kafka      Kafka      `yaml:"kafka"`
	Cache      Cache      `yaml:"cache"`
}

type HTTPServer struct {
	Port        string        `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type Database struct {
	DB_CONNECTION_STRING string `yaml:"db_connection_string"`
}

type Kafka struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
	Version string   `yaml:"version"`
}

type Cache struct {
	TTL time.Duration `yaml:"ttl"`
}

var (
	configPath           string
	DB_connection_string string
)

func NewConfig() *Config {
	var cfg Config

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	configPath = os.Getenv("CONFIG_PATH")
	DB_connection_string = os.Getenv("DB_CONNECTION_STRING")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if DB_connection_string == "" {
		log.Fatal("DB_connection_string is not set")
	}

	cfg.Database.DB_CONNECTION_STRING = DB_connection_string

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
