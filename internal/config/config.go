package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	Dsn        string     `yaml:"dsn"`
	PrivateKey string     `yaml:"private_key"`
	PublicKey  string     `yaml:"public_key"`
	HTTPServer HTTPServer `yaml:"http_server"`
	GRPCServer GRPCServer `yaml:"grpc_server"`
	Auth       Auth       `yaml:"auth"`
	Cache      Cache      `yaml:"cache"`
	Jaeger     Jaeger     `yaml:"jaeger"`
}

type GRPCServer struct {
	Port string `yaml:"port" env-default:"8080"`
}

type HTTPServer struct {
	Port        string        `yaml:"port" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idleTimeout" env-default:"60s"`
}

type Auth struct {
	PrivateKey     string        `yaml:"private_key" env-default:""`
	ExpirationTime time.Duration `yaml:"expiration_time"`
}

type Cache struct {
	ExpirationTime time.Duration `yaml:"expiration_time"`
	RedisAddr      string        `yaml:"redis_addr" env-default:"redis:6379"`
}

type Jaeger struct {
	Host        string `yaml:"host"`
	ServiceName string `yaml:"service_name"`
	LogSpans    bool   `yaml:"log_spans"`
}

func MustLoad() *Config {
	envFile, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	configPath := envFile["CONFIG_PATH"]
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %s", err)
	}
	return &cfg
}
