package main

import (
	"golf-game-kaffip/internal/bootstrap"
	"log"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file automatically when running locally
	godotenv.Load()
}

func main() {
	app, err := bootstrap.Initialize()
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	if err := bootstrap.Run(app); err != nil {
		log.Fatalf("failed to run application: %v", err)
	}
}
