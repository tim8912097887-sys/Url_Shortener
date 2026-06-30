package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configs struct{
	Addr string
	DbUrl string
	RedisAddr string
    RedisPassword string
	RedisDB int
}

func InitConfigs() (Configs, error) {
	_ = godotenv.Load()
	
	return Configs{
		Addr: getEnv("ADDR", ":8080"),
		DbUrl: getEnv("DB_URL","postgres://postgres:password@db:5432/url_shortener?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR","redis_server:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD","password"),
		RedisDB: getEnvFromInt("REDIS_DB",0),
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