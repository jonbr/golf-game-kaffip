package game

import (
	domainCourse "golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
)

// MapGameRowToDomain assembles a complete domain Game from its row plus
// the related course holes and hole results loaded separately. TeamA/TeamB
// are populated by the caller afterward via player repository lookups,
// since player data isn't available at this mapping layer.
func MapGameRowToDomain(row *GameRow, holes []CourseHoleRow, results []HoleResultRow, scoresByResult map[int64][]HoleResultScoreRow) *domainGame.Game {
	course := &domainCourse.Course{
		ID:    row.CourseID,
		Name:  row.CourseName,
		Holes: len(holes),
	}
	for _, h := range holes {
		course.HolesData = append(course.HolesData, domainCourse.Hole{
			Number:        h.HoleNumber,
			Par:           h.Par,
			HandicapIndex: h.HandicapIndex,
		})
	}

	holeResults := make(map[int]*domainGame.HoleResult, len(results))
	for _, r := range results {
		hr := &domainGame.HoleResult{
			Hole: domainGame.HoleInfo{
				Number: r.HoleNumber,
			},
			LowScoreWinnerTeam:  derefOrEmpty(r.LowScoreWinnerTeam),
			TeamTotalWinnerTeam: derefOrEmpty(r.TeamTotalWinnerTeam),
			PointsA:             r.PointsA,
			PointsB:             r.PointsB,
		}

		for _, s := range scoresByResult[r.ID] {
			hr.Scores = append(hr.Scores, domainGame.PlayerHoleResult{
				PlayerID:   s.PlayerID,
				Gross:      s.Gross,
				Net:        s.Net,
				Strokes:    s.Strokes,
				GrossBonus: s.GrossBonus,
			})
			if s.GrossBonus > 0 {
				hr.GrossBonuses = append(hr.GrossBonuses, domainGame.GrossBonus{
					PlayerID: s.PlayerID,
					Bonus:    s.GrossBonus,
				})
			}
		}

		holeResults[r.HoleNumber] = hr
	}

	return &domainGame.Game{
		ID:           row.ID,
		Course:       course,
		Variant:      domainGame.Variant(row.Variant),
		StartingLead: row.StartingLead,
		CurrentHole:  row.CurrentHole,
		MatchScore: domainGame.MatchScore{
			TeamA: row.MatchTeamA,
			TeamB: row.MatchTeamB,
		},
		HoleResults: holeResults,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
		FinishedAt:  row.FinishedAt,
	}
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func MapGameSummaryRowToDomain(row *GameSummaryRow) *domainGame.GameSummary {
	return &domainGame.GameSummary{
		ID:          row.ID,
		CourseID:    row.CourseID,
		CourseName:  row.CourseName,
		Variant:     domainGame.Variant(row.Variant),
		CurrentHole: row.CurrentHole,
		TotalHoles:  row.TotalHoles,
		MatchScore: domainGame.MatchScore{
			TeamA: row.MatchTeamA,
			TeamB: row.MatchTeamB,
		},
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		FinishedAt: row.FinishedAt,
	}
}
