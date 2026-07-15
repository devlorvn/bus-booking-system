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
	BookingLockTTL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil && os.Getenv("APP_MODE") != "production" {
		panic("Error loading .env file")
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
		BookingLockTTL: getEnv("BOOKING_LOCK_TTL"),
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Environment variable %s not set", key))
	}
	return value
}
