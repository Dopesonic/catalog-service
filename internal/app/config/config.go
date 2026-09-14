package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/Dopesonic/catalog-service/internal/app/config/section"
)

type Config struct {
	Repository section.Repository
	Processor  section.Processor
	Monitor    section.Monitor
}

var Root Config

func Load() {
	_ = godotenv.Load()

	err := envconfig.Process("APP", &Root)
	if err != nil {
		log.Fatal(err)
	}
}
