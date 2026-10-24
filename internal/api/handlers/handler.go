package handlers

import (
	"golf-game-kaffip/internal/api/middleware"
	"golf-game-kaffip/internal/application"
	"golf-game-kaffip/internal/games"
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
	Games               []games.Module
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
	gameModules []games.Module,
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
		Games:               gameModules,
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

	// Cups
	r.Post("/cups/match_play", h.CreateMatchPlayCup)
	r.Get("/cups/{id}", h.GetCup)

	// Games (shared across formats)
	r.Get("/games", h.GetGames)
	r.Post("/games/{id}/finish", h.FinishGame)

	// Games not yet migrated to their own module
	r.Get("/games", h.GetGames)
	r.Post("/games/match_play", h.CreateMatchPlay)
	r.Post("/games/team_points", h.CreateTeamPoints)
	r.Post("/games/wolf_play", h.CreateWolfPlay)
	//r.Post("/games/alternate_shot", h.CreateAlternateShot)
	r.Get("/games/match_play/{id}", h.GetMatchPlayGame)
	r.Get("/games/team_points/{id}", h.GetTeamPointsGame)
	r.Get("/games/wolf/{id}", h.GetWolfPlay)
	r.Put("/games/{id}/holes/{holeNumber}/score", h.SetHoleScore)
	r.Post("/games/{id}/finish", h.FinishGame)

	// Games migrated to modules (one module per format: create))
	for _, m := range h.Games {
		m.RegisterRoutes(r)
	}

	// Courses External API
	r.Get("/courses/search", h.SearchCourses)

	// Print all registered routes
	PrintRoutes(r)

	return r
}
