package handlers

import (
	"context"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/middleware"
	"golf-game-kaffip/internal/application"
	"golf-game-kaffip/internal/logging"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	CupService          *application.CupService
	GameService         *application.GameService
	MatchPlayService    *application.MatchPlayService
	TeamPointsService   *application.TeamPointsService
	WolfGameService     *application.WolfGameService
	PlayerService       *application.PlayerService
	MatchPlayCupService *application.MatchPlayCupService
	Logger              *slog.Logger
	DB                  *pgxpool.Pool
	CORSAllowedOrigins  []string
}

func NewHandler(
	cupService *application.CupService,
	gameService *application.GameService,
	matchPlayService *application.MatchPlayService,
	teamPointsService *application.TeamPointsService,
	wolfGameService *application.WolfGameService,
	playerService *application.PlayerService,
	matchPlayCupService *application.MatchPlayCupService,
	logger *slog.Logger,
	db *pgxpool.Pool,
	corsAllowedOrigins []string,
) *Handler {
	return &Handler{
		CupService:          cupService,
		GameService:         gameService,
		MatchPlayService:    matchPlayService,
		TeamPointsService:   teamPointsService,
		WolfGameService:     wolfGameService,
		PlayerService:       playerService,
		MatchPlayCupService: matchPlayCupService,
		Logger:              logger,
		DB:                  db,
		CORSAllowedOrigins:  corsAllowedOrigins,
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.CORSMiddleware(h.CORSAllowedOrigins))
	r.Use(middleware.LoggerMiddleware(h.Logger))
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
	r.Post("/cups/match_play", h.CreateMatchPlayCup)
	r.Post("/games/team_points", h.CreateTeamPoints)
	r.Post("/games/match_play", h.CreateMatchPlay)
	r.Post("/games/wolf_play", h.CreateWolfPlay)

	r.Get("/games", h.GetGames)
	r.Get("/cups/{id}", h.GetCup)
	r.Get("/games/team_points/{id}", h.GetTeamPointsGame)
	r.Get("/games/match_play/{id}", h.GetMatchPlayGame)
	r.Get("/games/wolf/{id}", h.GetWolfPlay)

	r.Put("/games/{id}/holes/{holeNumber}/score", h.SetHoleScore)

	r.Post("/games/{id}/finish", h.FinishGame)

	// Courses External API
	r.Get("/courses/search", h.SearchCourses)

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

// TODO: Make function agnostic so we can parse other Id's as well.
func parseID(w http.ResponseWriter, r *http.Request, logger *slog.Logger, entityName string) (string, bool) {
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Error("missing " + entityName + " id")
		api.WriteBadRequest(w, "missing_"+entityName+"_id", entityName+" id must be set", nil)
		return "", false
	}
	return id, true
}
