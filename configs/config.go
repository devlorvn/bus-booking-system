package configs

import (
	"fmt"
	"os"

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

type Config struct {
	Port string
	Mode string
	Database
	Redis
	BookingLockTTL string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

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
