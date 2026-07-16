package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"golf-game-kaffip/internal/api/handlers"
	"golf-game-kaffip/internal/application"
	"golf-game-kaffip/internal/config"
	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
	"log/slog"

	gamedb "golf-game-kaffip/internal/infrastructure/postgres/game"
	playerdb "golf-game-kaffip/internal/infrastructure/postgres/player"
	"net/http"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config      config.Config
	Logger      *slog.Logger
	DB          *pgxpool.Pool
	GameService *application.GameService
	Handlers    *handlers.Handler
	Router      http.Handler
}

func Initialize() (*App, error) {
	// -----------------------------
	// Load config and logger
	// -----------------------------
	cfg := config.Load()
	logger := config.GetLogger()

	// -----------------------------
	// 1. Connect to Postgres
	// -----------------------------
	dsn := os.Getenv("DATABASE_URL")
	// Ensure DB exists BEFORE migrations
	if err := ensureDatabaseExists(dsn); err != nil {
		return nil, err
	}

	// Run migrations BEFORE creating pgxpool
	if err := runMigrations(dsn); err != nil {
		return nil, err
	}

	// Now create pgxpool for the app
	db, err := connectPostgres()
	if err != nil {
		return nil, err
	}

	// -----------------------------
	// 1. Instantiate external API client
	// -----------------------------
	externalAPI := opengolfapi.NewClient(cfg.OpenGolfAPIKey, logger)

	// -----------------------------
	// 2. Instantiate repositories
	// -----------------------------
	var (
		playerRepo player.Repository = playerdb.NewRepository(db)
		gameRepo   game.Repository   = gamedb.NewRepository(db, playerRepo)
	)

	// -----------------------------
	// 3. Create application services
	// -----------------------------
	gameService := application.NewGameService(
		gameRepo,
		playerRepo,
		externalAPI,
	)

	playerService := application.NewPlayerService(playerRepo)

	// -----------------------------
	// 4. Create HTTP handlers
	// -----------------------------
	h := handlers.NewHandler(gameService, playerService, logger, db)

	return &App{
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		GameService: gameService,
		Handlers:    h,
		Router:      h.Router(),
	}, nil
}

func runMigrations(dsn string) error {
	// Open a dedicated sql.DB for migrations
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open sql.DB: %w", err)
	}
	defer db.Close()

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/postgres/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	m.Log = migrateLogger{}

	// Apply migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

func connectPostgres() (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}

	// Optional tuning (good defaults for API workloads)
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	// Verify connection works
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return pool, nil
}

func (a *App) Close() error {
	if a.DB != nil {
		a.DB.Close() // pgxpool.Pool.Close() has no return value
	}
	return nil
}
