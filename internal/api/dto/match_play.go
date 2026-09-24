package dto

// TODO: Refactore this struct, team is not needed as this is
// indiviually played variant.
type CreateMatchPlayRequest struct {
	CourseID string  `json:"course_id"`
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
	Variant  string  `json:"variant"` // "gross" or "net"
}
