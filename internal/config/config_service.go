package config

import (
	"log"
	"os"
	"gopkg.in/yaml.v3"
)


type Config struct{
	Env string `yaml:"env" env-default:"local"`
	HttpPort int `yaml:"http_port" env-default:"8080"`
	DbName string `yaml:"db_name"`
	DbUser string `yaml:"db_user"`
	DbPassword string `yaml:"db_password"`
	DbUrl string `yaml:"db_url"`
	DbPort string `yaml:"db_port"`
	CacheSize int `yaml:"cache_size" env-default:"10"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoad() *Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		log.Fatal("CONFIG_PATH env is required")
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return cfg
}