package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SslMode  string
}
type Redis struct {
	Host string
	Port string
}

type Kafka struct {
	Brokers []string
}

type Config struct {
	Port string
	Mode string
	Database
	Redis
	Kafka
	BookingLockTTL     string
	BusServiceAddr     string
	BookingServiceAddr string
	UserServiceAddr    string
	MetricsPort        string
}

func LoadConfig() *Config {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			panic("Error loading .env file")
		}
	}

	// read list brokers
	brokerStr := getEnv("KAFKA_BROKERS")
	brokers := strings.Split(brokerStr, ",")

	return &Config{
		Port: getEnv("APP_PORT"),
		Mode: getEnv("APP_MODE"),
		Database: Database{
			Host:     getEnv("POSTGRES_HOST"),
			Port:     getEnv("POSTGRES_PORT"),
			User:     getEnv("POSTGRES_USER"),
			Password: getEnv("POSTGRES_PASSWORD"),
			Name:     getEnv("POSTGRES_DB"),
			SslMode:  getEnv("POSTGRES_SSLMODE"),
		},
		Redis: Redis{
			Host: getEnv("REDIS_HOST"),
			Port: getEnv("REDIS_PORT"),
		},
		Kafka: Kafka{
			Brokers: brokers,
		},
		BookingLockTTL:     getEnv("BOOKING_LOCK_TTL"),
		BusServiceAddr:     getEnvOrDefault("BUS_SERVICE_ADDR", "127.0.0.1:50051"),
		BookingServiceAddr: getEnvOrDefault("BOOKING_SERVICE_ADDR", "127.0.0.1:50052"),
		UserServiceAddr:    getEnvOrDefault("USER_SERVICE_ADDR", "127.0.0.1:50053"),
		MetricsPort:        getEnvOrDefault("METRICS_PORT", "9090"),
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s not set", key))
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
