package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Configs struct{
	Addr string
	DbUrl string
}

func InitConfigs() (Configs, error) {
	godotenv.Load()
    if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
        return Configs{}, err
    }
	return Configs{
		Addr: getEnv("ADDR", ":8080"),
		DbUrl: getEnv("DB_URL","postgres://postgres:password@db:5432/url_shortener?sslmode=disable"),
	},nil
}

func getEnv(key string, defaultValue string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}