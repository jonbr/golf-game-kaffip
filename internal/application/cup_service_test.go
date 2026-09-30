// internal/application/cup_service_test.go
package application

import (
	"context"
	"testing"
	"time"

	"golf-game-kaffip/internal/domain/cup"
	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/domain/wolf"
)

type fakeCupRepository struct {
	cup *cup.Cup
}

func (f *fakeCupRepository) CreateCup(ctx context.Context, c *cup.Cup) error { return nil }
func (f *fakeCupRepository) LoadCup(ctx context.Context, id string) (*cup.Cup, error) {
	return f.cup, nil
}
func (f *fakeCupRepository) FinishCup(ctx context.Context, id string) error { return nil }

type fakeGameRepository struct {
	types map[string]game.GameType
	games map[string]*game.Game
}

func (f *fakeGameRepository) GetGameType(ctx context.Context, id string) (game.GameType, error) {
	return f.types[id], nil
}
func (f *fakeGameRepository) LoadGame(ctx context.Context, id string) (*game.Game, error) {
	return f.games[id], nil
}
func (f *fakeGameRepository) CreateGame(ctx context.Context, g *game.Game) error { return nil }
func (f *fakeGameRepository) ListSummaries(ctx context.Context, opts game.ListOptions) ([]*game.GameSummary, error) {
	return nil, nil
}
func (f *fakeGameRepository) SaveHoleResult(ctx context.Context, g *game.Game, holeNumber int) error {
	return nil
}
func (f *fakeGameRepository) FinishGame(ctx context.Context, id string) error { return nil }
func (f *fakeGameRepository) PlayersInActiveGame(ctx context.Context, playerIDs []int64) (int64, error) {
	return 0, nil
}
func (f *fakeGameRepository) PlayerExists(ctx context.Context, id int64) (bool, error) {
	return true, nil
}

type fakeWolfRepository struct {
	games map[string]*wolf.Game
}

func (f *fakeWolfRepository) CreateGame(ctx context.Context, g *wolf.Game) error { return nil }
func (f *fakeWolfRepository) LoadGame(ctx context.Context, id string) (*wolf.Game, error) {
	return f.games[id], nil
}
func (f *fakeWolfRepository) SaveHoleResult(ctx context.Context, g *wolf.Game, holeNumber int) error {
	return nil
}

func mkP(id int64) *player.Player { return &player.Player{ID: id} }

func TestCupService_GetCup_AggregatesAcrossFormats(t *testing.T) {
	roster := map[int64]cup.Side{
		1: cup.SideA, 2: cup.SideB, // match play
		3: cup.SideA, 4: cup.SideB, 5: cup.SideA, 6: cup.SideB, // wolf
	}

	c := &cup.Cup{
		ID:       "cup_1",
		Players:  roster,
		MatchIDs: []string{"game_mp", "game_wolf"},
	}

	matchPlayGame := &game.Game{
		ID:         "game_mp",
		TeamA:      []*player.Player{mkP(1)},
		TeamB:      []*player.Player{mkP(2)},
		MatchScore: game.MatchScore{TeamA: 1, TeamB: 0},
		FinishedAt: timePtr(),
	}

	wolfGame := &wolf.Game{
		ID:         "game_wolf",
		Players:    [4]*player.Player{mkP(3), mkP(4), mkP(5), mkP(6)},
		FinishedAt: timePtr(),
		HoleResults: map[int]*wolf.HoleResult{
			1: {PointsAwarded: map[int64]int{3: 3, 4: 0, 5: 0, 6: 0}}, // player 3 (side A) wins outright
		},
	}

	svc := NewCupService(
		&fakeCupRepository{cup: c},
		&fakeGameRepository{
			types: map[string]game.GameType{"game_mp": game.GameTypeMatchPlay, "game_wolf": game.GameTypeWolf},
			games: map[string]*game.Game{"game_mp": matchPlayGame},
		},
		&fakeWolfRepository{games: map[string]*wolf.Game{"game_wolf": wolfGame}},
	)

	detail, err := svc.GetCup(context.Background(), "cup_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// match_play: A wins (1-0). wolf: A wins (1-0). Total: A=2, B=0.
	if detail.Score.TeamA != 2 || detail.Score.TeamB != 0 {
		t.Errorf("Score = %+v, want {TeamA:2 TeamB:0}", detail.Score)
	}
}

func timePtr() *time.Time { t := time.Now(); return &t }
