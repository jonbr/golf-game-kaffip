package application

import (
	"context"
	"errors"
	"testing"

	"golf-game-kaffip/internal/domain/player"
)

// fakePlayerRepository is a minimal in-memory stand-in for player.Repository,
// just enough to test PlayerService's own logic without a real DB.
type fakePlayerRepository struct {
	createErr    error
	createCalled bool
}

func (f *fakePlayerRepository) Create(ctx context.Context, p *player.Player) error {
	f.createCalled = true
	if f.createErr != nil {
		return f.createErr
	}
	p.ID = 1
	return nil
}

func (f *fakePlayerRepository) FindAll(ctx context.Context) ([]*player.Player, error) {
	return nil, nil
}
func (f *fakePlayerRepository) FindByID(ctx context.Context, id int64, includeDeleted bool) (*player.Player, error) {
	return nil, nil
}
func (f *fakePlayerRepository) Update(ctx context.Context, p *player.Player) error { return nil }
func (f *fakePlayerRepository) SoftDelete(ctx context.Context, id int64) error     { return nil }
func (f *fakePlayerRepository) GetActiveGameForPlayer(ctx context.Context, id int64) (*string, error) {
	return nil, nil
}
func (f *fakePlayerRepository) GetTeamsForGame(ctx context.Context, gameID string) ([]*player.Player, []*player.Player, error) {
	return nil, nil, nil
}

func TestPlayerService_CreatePlayer_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		pName    string
		email    string
		handicap float64
		wantCode string
	}{
		{"empty name rejected", "", "a@b.com", 10, "validation_error"},
		{"empty email rejected", "Alice", "", 10, "validation_error"},
		{"negative handicap rejected", "Alice", "a@b.com", -1, "validation_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakePlayerRepository{}
			s := NewPlayerService(repo)

			_, err := s.CreatePlayer(context.Background(), tt.pName, tt.email, tt.handicap)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			var svcErr *ServiceError
			if !errors.As(err, &svcErr) {
				t.Fatalf("expected *ServiceError, got %T: %v", err, err)
			}
			if svcErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", svcErr.Code, tt.wantCode)
			}
			if repo.createCalled {
				t.Error("expected repository.Create to never be called for invalid input")
			}
		})
	}
}

func TestPlayerService_CreatePlayer_Success(t *testing.T) {
	repo := &fakePlayerRepository{}
	s := NewPlayerService(repo)

	p, err := s.CreatePlayer(context.Background(), "Alice", "alice@example.com", 12.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != 1 {
		t.Errorf("expected repository-assigned ID 1, got %d", p.ID)
	}
	if !repo.createCalled {
		t.Error("expected repository.Create to be called")
	}
}

func TestPlayerService_CreatePlayer_DuplicateEmailTranslated(t *testing.T) {
	repo := &fakePlayerRepository{createErr: player.ErrEmailAlreadyExists}
	s := NewPlayerService(repo)

	_, err := s.CreatePlayer(context.Background(), "Alice", "alice@example.com", 12.5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var svcErr *ServiceError
	if !errors.As(err, &svcErr) {
		t.Fatalf("expected *ServiceError, got %T: %v", err, err)
	}
	if svcErr.Code != "email_already_exists" {
		t.Errorf("Code = %q, want email_already_exists", svcErr.Code)
	}
}
