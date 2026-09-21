package application

import (
	"context"
	"fmt"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/game"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
	"golf-game-kaffip/internal/logging"
	"time"
)

type GameService struct {
	games                 game.Repository
	players               player.Repository
	externalCourseService *ExternalCourseService
}

func NewGameService(
	games game.Repository,
	players player.Repository,
	externalAPI opengolfapi.ClientInterface,
) *GameService {
	return &GameService{
		games:                 games,
		players:               players,
		externalCourseService: NewExternalCourseService(externalAPI),
	}
}

func (s *GameService) CreateGame(ctx context.Context, gameType domainGame.GameType, req dto.CreateGameRequest) (*domainGame.Game, error) {
	logger := logging.FromCtx(ctx)

	if err := validateTeamSize(gameType, req.TeamA, req.TeamB); err != nil {
		return nil, err
	}

	playersIDs := append(req.TeamA, req.TeamB...)

	if err := validatePlayersExist(ctx, s.games, logger, playersIDs); err != nil {
		return nil, err
	}
	if err := validateActiveGameConflict(ctx, s.games, logger, playersIDs); err != nil {
		return nil, err
	}

	teamAPlayers, err := loadPlayers(ctx, s.players, req.TeamA)
	if err != nil {
		return nil, err
	}
	teamBPlayers, err := loadPlayers(ctx, s.players, req.TeamB)
	if err != nil {
		return nil, err
	}

	course, err := fetchCourse(ctx, s.externalCourseService, logger, req.CourseID)
	if err != nil {
		return nil, err
	}

	gameID := fmt.Sprintf("game_%d", time.Now().UnixNano())

	// 5. Create domain game
	g, err := domainGame.NewGame(gameID, course, teamAPlayers, teamBPlayers, domainGame.GameType(gameType), domainGame.Variant(req.Variant))
	if err != nil {
		return nil, NewServiceError("invalid_game_params", map[string]any{"underlying": err.Error()})
	}

	// 4. Persist
	if err := s.games.CreateGame(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	return g, nil
}

func (s *GameService) GetGames(ctx context.Context, status string) ([]*domainGame.GameSummary, error) {
	opts := domainGame.ListOptions{}
	switch status {
	case "", "all":
	case "active":
		opts.ActiveOnly = true
	case "finished":
		opts.FinishedOnly = true
	default:
		return nil, NewServiceError("invalid_status_filter", map[string]any{"status": status})
	}

	rows, err := s.games.ListSummaries(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve games: %w", err)
	}

	return rows, nil
}

func (s *GameService) GetGame(ctx context.Context, id string) (*domainGame.Game, error) {
	// LoadGame already returns a fully populated Game (course holes,
	// GameType, Variant, StartingLead, HoleResults, TeamA/TeamB) straight
	// from local storage — no need to rebuild it or hit the external
	// course API again (that only happens once, at CreateGame).
	g, err := s.games.LoadGame(ctx, id)
	if err != nil {
		return nil, NewServiceError("game_not_found", map[string]any{"game_id": id})
	}

	return g, nil
}

func (s *GameService) SearchCourses(ctx context.Context, query string) ([]opengolfapi.CourseSearchResult, error) {
	results, err := s.externalCourseService.SearchCourses(ctx, query)
	if err != nil {
		return nil, NewServiceError("external_api_error", map[string]any{"underlying": err.Error()})
	}
	return results, nil
}

func (s *GameService) SetHoleScore(ctx context.Context, gameID string, holeNumber int, scores []dto.PlayerGrossScore) (*domainGame.Game, error) {
	logger := logging.FromCtx(ctx)

	g, err := s.games.LoadGame(ctx, gameID)
	if err != nil {
		return nil, NewServiceError("game_not_found", map[string]any{"game_id": gameID})
	}

	inputs, err := buildScoreInputs(g, scores)
	if err != nil {
		return nil, err
	}

	if err := g.SetHoleScore(holeNumber, inputs); err != nil {
		logger.Error("failed to set hole score", "game_id", gameID, "hole_number", holeNumber, "error", err)
		return nil, NewServiceError("invalid_hole_score", map[string]any{
			"game_id":     gameID,
			"hole_number": holeNumber,
			"underlying":  err.Error(),
		})
	}

	if err := s.games.SaveHoleResult(ctx, g, holeNumber); err != nil {
		return nil, fmt.Errorf("failed to persist hole %d result: %w", holeNumber, err)
	}

	return g, nil
}

func (s *GameService) FinishGame(ctx context.Context, gameID string) error {
	g, err := s.GetGame(ctx, gameID)
	if err != nil {
		return err
	}

	if g.FinishedAt != nil {
		return NewServiceError("game_already_finished", map[string]any{"game_id": gameID})
	}

	return s.games.FinishGame(ctx, gameID)
}

// buildScoreInputs validates that scores are provided for exactly the
// game's registered players (4 for team play, 2 for match play) and tags
// each with its TeamID.
func buildScoreInputs(g *domainGame.Game, scores []dto.PlayerGrossScore) ([]domainGame.PlayerScoreInput, error) {
	expected := 4
	if g.GameType == domainGame.GameTypeMatchPlay {
		expected = 2
	}
	if len(scores) != expected {
		return nil, NewServiceError("invalid_score_count", map[string]any{
			"expected": expected,
			"got":      len(scores),
		})
	}

	teamOf := make(map[int64]string, 4)
	for _, p := range g.TeamA {
		teamOf[p.ID] = "A"
	}
	for _, p := range g.TeamB {
		teamOf[p.ID] = "B"
	}

	inputs := make([]domainGame.PlayerScoreInput, 0, 4)
	seen := make(map[int64]bool, 4)

	for _, s := range scores {
		team, ok := teamOf[s.PlayerID]
		if !ok {
			return nil, NewServiceError("player_not_in_game", map[string]any{
				"player_id": s.PlayerID,
				"game_id":   g.ID,
			})
		}
		if seen[s.PlayerID] {
			return nil, NewServiceError("duplicate_player_score", map[string]any{
				"player_id": s.PlayerID,
			})
		}
		seen[s.PlayerID] = true

		inputs = append(inputs, domainGame.PlayerScoreInput{
			PlayerID: s.PlayerID,
			Gross:    s.Gross,
			TeamID:   team,
		})
	}

	return inputs, nil
}

/*func (s *GameService) loadPlayers(ctx context.Context, ids []int64) ([]*player.Player, error) {
	players := make([]*player.Player, 0, len(ids))
	for _, id := range ids {
		p, err := s.players.FindByID(ctx, id, false)
		if err != nil {
			return nil, fmt.Errorf("failed to load player with ID %d: %w", id, err)
		}
		players = append(players, p)
	}

	return players, nil
}*/

func validateTeamSize(gameType domainGame.GameType, teamA, teamB []int64) error {
	var want int
	switch gameType {
	case domainGame.GameTypePointsPlay:
		want = 2
	case domainGame.GameTypeMatchPlay:
		want = 1
	default:
		return NewServiceError("invalid_game_type", map[string]any{"game_type": gameType})
	}

	if len(teamA) != want || len(teamB) != want {
		return NewServiceError("invalid_team_size", map[string]any{
			"game_type":         gameType,
			"expected_per_side": want,
			"team_a_size":       len(teamA),
			"team_b_size":       len(teamB),
		})
	}
	return nil
}
