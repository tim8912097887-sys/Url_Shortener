package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configs struct{
	Addr string
	ClientOrigin string
	DbUrl string
	RedisUrl string
}

func InitConfigs() (Configs, error) {
	_ = godotenv.Load()
	
	return Configs{
		Addr: getEnv("ADDR", ":8080"),
		ClientOrigin: getEnv("CLIENT_ORIGIN", "http://localhost:5173"),
		DbUrl: getEnv("DB_URL","postgres://postgres:password@db:5432/url_shortener?sslmode=disable"),
		RedisUrl: getEnv("REDIS_URL","redis://redis:6379"),
	},nil
}

func getEnv(key string, defaultValue string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func getEnvFromInt(key string, defaultValue int) int {

	if value, ok := os.LookupEnv(key); ok {
		num, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return num
	}

	return defaultValue
}