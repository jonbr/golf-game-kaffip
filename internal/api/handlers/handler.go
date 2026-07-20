package handlers

import (
	"context"
	"golf-game-kaffip/internal/api/middleware"
	"golf-game-kaffip/internal/application"
	"golf-game-kaffip/internal/logging"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	GameService        *application.GameService
	PlayerService      *application.PlayerService
	Logger             *slog.Logger
	DB                 *pgxpool.Pool
	CORSAllowedOrigins []string
}

func NewHandler(
	gameService *application.GameService,
	playerService *application.PlayerService,
	logger *slog.Logger,
	db *pgxpool.Pool,
	corsAllowedOrigins []string,
) *Handler {
	return &Handler{
		GameService:        gameService,
		PlayerService:      playerService,
		Logger:             logger,
		DB:                 db,
		CORSAllowedOrigins: corsAllowedOrigins,
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.CORSMiddleware(h.CORSAllowedOrigins))
	r.Use(middleware.LoggerMiddleware(h.Logger))
	r.Use(middleware.RequestIDMiddleware(h.Logger))
	r.Use(middleware.RecoveryMiddleware(h.Logger))

	// health check
	r.Get("/health", h.Health)

	// Players
	r.Post("/players", h.CreatePlayer)
	r.Get("/players", h.GetPlayers)
	r.Get("/players/{id}", h.GetPlayer)
	r.Put("/players/{id}", h.UpdatePlayer)
	r.Delete("/players/{id}", h.DeletePlayer)

	// Games
	r.Post("/games/team_play", h.CreateGame)
	r.Post("/games/match_play", h.CreateGame)
	r.Get("/games", h.GetGames)
	r.Get("/games/{id}", h.GetGame)
	r.Put("/games/{id}/holes/{holeNumber}/score", h.SetHoleScore)
	r.Post("/games/{id}/finish", h.FinishGame)

	// Print all registered routes
	PrintRoutes(r)

	return r
}

func startRequest(r *http.Request, action string) (context.Context, *slog.Logger) {
	ctx := r.Context()
	logger := logging.FromCtx(ctx)
	logger.Info(action)
	return ctx, logger
}
