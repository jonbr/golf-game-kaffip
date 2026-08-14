package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"golf-game-kaffip/internal/bootstrap"
)

func init() {
	godotenv.Load()
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	if err := bootstrap.RunMigrations(dsn); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations complete")
}
