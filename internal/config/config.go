package config

import (
	"os"
	"strconv"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	JWTSecret              string
	JWT_ACCESS_EXPIRATION  int
	JWT_REFRESH_EXPIRATION int
	JWT_2FA_EXPIRATION     int
	JWT_DOMAIN             string
	JWT_PATH               string
}

type Config struct {
	NODE_ENV          string
	Port              string
	CLIENT_URL        string
	GATEWAY_API_TOKEN string
	JWT               JWTConfig
	Database          DatabaseConfig
	Redis             RedisConfig
}

func Load() (*Config, error) {
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDB = 0
	}

	return &Config{
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
		Database: DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		},
		NODE_ENV:          os.Getenv("NODE_ENV"),
		Port:              os.Getenv("PORT"),
		CLIENT_URL:        os.Getenv("CLIENT_URL"),
		GATEWAY_API_TOKEN: os.Getenv("GATEWAY_API_TOKEN"),
		JWT: JWTConfig{
			JWTSecret: os.Getenv("JWT_SECRET"),
			JWT_ACCESS_EXPIRATION: func() int {
				val, err := strconv.Atoi(os.Getenv("JWT_ACCESS_EXPIRATION"))
				if err != nil {
					return 0
				}
				return val
			}(),
			JWT_REFRESH_EXPIRATION: func() int {
				val, err := strconv.Atoi(os.Getenv("JWT_REFRESH_EXPIRATION"))
				if err != nil {
					return 0
				}
				return val
			}(),
			JWT_2FA_EXPIRATION: func() int {
				val, err := strconv.Atoi(os.Getenv("JWT_2FA_EXPIRATION"))
				if err != nil {
					return 0
				}
				return val
			}(),
			JWT_DOMAIN: os.Getenv("JWT_DOMAIN"),
			JWT_PATH:   os.Getenv("JWT_PATH"),
		},
	}, nil
}
