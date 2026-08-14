package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainCourse "golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
)

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create event")

	var req dto.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	event, err := h.TeamEventService.CreateEvent(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create event failed", "error", err)
		api.WriteError(w, err)
		return
	}

	matchIDs := make([]string, len(event.Matches))
	for i, m := range event.Matches {
		matchIDs[i] = m.ID
	}

	api.JSON(w, http.StatusCreated, dto.CreateEventResponse{
		EventID:  event.ID,
		MatchIDs: matchIDs,
	})
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get event")

	id := chi.URLParam(r, "id")
	if id == "" {
		api.WriteBadRequest(w, "missing_event_id", "event id must be set", nil)
		return
	}

	event, err := h.TeamEventService.GetEvent(ctx, id)
	if err != nil {
		logger.Error("get event failed", "event_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, mapEventToResponse(event))
}

func (h *Handler) FinishEvent(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "finish event")

	id := chi.URLParam(r, "id")
	if id == "" {
		api.WriteBadRequest(w, "missing_event_id", "event id must be set", nil)
		return
	}

	if err := h.TeamEventService.FinishEvent(ctx, id); err != nil {
		logger.Error("finish event failed", "event_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapEventToResponse(event *domainGame.TeamEvent) dto.EventResponse {
	score := event.Score()

	matches := make([]dto.EventMatchResponse, len(event.Matches))
	for i, m := range event.Matches {
		outcome := m.Outcome()

		teamA := make([]dto.PlayerSummaryResponse, len(m.TeamA))
		for j, p := range m.TeamA {
			teamA[j] = dto.PlayerSummaryResponse{ID: p.ID, Name: p.Name, Handicap: p.Handicap}
		}
		teamB := make([]dto.PlayerSummaryResponse, len(m.TeamB))
		for j, p := range m.TeamB {
			teamB[j] = dto.PlayerSummaryResponse{ID: p.ID, Name: p.Name, Handicap: p.Handicap}
		}

		matches[i] = dto.EventMatchResponse{
			GameID:      m.ID,
			GameType:    string(m.GameType),
			TeamA:       teamA,
			TeamB:       teamB,
			CurrentHole: m.CurrentHole,
			Status: dto.MatchOutcomeResponse{
				Display: outcome.Display, WinnerTeam: outcome.WinnerTeam, Decided: outcome.Decided,
			},
		}
	}

	return dto.EventResponse{
		ID:         event.ID,
		Course:     dto.CourseSummaryResponse{ID: event.Course.ID, Name: event.Course.Name},
		Variant:    string(event.Variant),
		Score:      dto.EventScoreResponse{TeamA: score.TeamA, TeamB: score.TeamB},
		Matches:    matches,
		FinishedAt: event.FinishedAt,
	}
}
