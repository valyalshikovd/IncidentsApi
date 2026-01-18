package configs

import (
	"fmt"
	"os"
)

type DbConfig struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	DBTimeZone string
}

func Load() DbConfig {
	return DbConfig{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "postgres"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
	}
}

func (c DbConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s",
		c.DBHost, c.DBPort, c.DBName, c.DBUser, c.DBPassword,
	)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
