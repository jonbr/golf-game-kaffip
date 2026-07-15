package application

import (
	domainError "golf-game-kaffip/internal/domain/game"
)

type ServiceError struct {
	Code    string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func NewServiceError(code string, details any) *ServiceError {
	return &ServiceError{
		Code:    code,
		Details: details,
	}
}

func NewExternalServiceError(err error, details any) *ServiceError {
	return NewServiceError("external_api_error", details)
}

func (e *ServiceError) ServiceCode() string {
	return e.Code
}

func (e *ServiceError) Error() string {
	return e.Code
}

func MapDomainError(err error) *ServiceError {
	switch e := err.(type) {

	case domainError.PlayersInActiveGame:
		return NewServiceError("player_already_in_active_game", map[string]any{
			"player_id": e.PlayerID,
			"game_id":   e.GameID,
		})

	case domainError.ErrPlayerNotInGame:
		return NewServiceError("player_not_in_game", map[string]any{
			"player_id": e.PlayerID,
			"game_id":   e.GameID,
		})

	case domainError.ErrInvalidGrossScore:
		return NewServiceError("invalid_gross_score", map[string]any{
			"player_id": e.PlayerID,
			"game_id":   e.GameID,
			"gross":     e.Gross,
		})

	case domainError.ErrHoleOutOfRange:
		return NewServiceError("hole_out_of_range", map[string]any{
			"hole":    e.Hole,
			"game_id": e.GameID,
		})

	case domainError.ErrWrongHole:
		return NewServiceError("wrong_hole", map[string]any{
			"expected": e.Expected,
			"got":      e.Got,
			"game_id":  e.GameID,
		})

	case domainError.ErrGameFinished:
		return NewServiceError("game_finished", map[string]any{
			"game_id": e.GameID,
		})
	}

	return nil
}
