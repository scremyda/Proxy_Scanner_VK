package conf

import (
	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload" // Load enviroment from .env
	"log"
	"os"
)

type Config struct {
	DBName string `env:"POSTGRES_DB" env-default:"postgres"`
	DBPass string `env:"POSTGRES_PASSWORD" env-default:"1"`
	DBHost string `env:"DB_HOST" env-default:"127.0.0.1"`
	DBPort int    `env:"DB_PORT" env-default:"5432"`
	DBUser string `env:"POSTGRES_USER" env-default:"postgres"`
}

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Printf("cannot read .env file: %s\n (fix: you need to put .env file in main dir)", err)
		os.Exit(1)
	}

	return &cfg
}
