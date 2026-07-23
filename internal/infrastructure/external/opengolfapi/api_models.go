package opengolfapi

// internal/infrastructure/external/opengolf/api_models.go
type CourseResponse struct {
	ID             string     `json:"id"`
	CourseName     string     `json:"course_name"`
	ClubName       string     `json:"club_name"`
	City           string     `json:"city"`
	State          string     `json:"state"`
	Lat            float64    `json:"lat"`
	Lng            float64    `json:"lng"`
	Type           string     `json:"type"`
	Par            int        `json:"par"`
	Holes          int        `json:"holes"`
	Yardage        int        `json:"yardage"`
	Timezone       string     `json:"timezone"`
	Architect      any        `json:"architect"`
	YearBuilt      int        `json:"year_built"`
	Description    any        `json:"description"`
	DescriptionSrc any        `json:"description_source"`
	FairwayGrass   any        `json:"fairway_grass"`
	GreenGrass     any        `json:"green_grass"`
	Phone          any        `json:"phone"`
	Website        any        `json:"website"`
	Address        any        `json:"address"`
	PostalCode     any        `json:"postal_code"`
	Facilities     any        `json:"facilities"`
	Ratings        any        `json:"ratings"`
	Sources        any        `json:"sources"`
	Tees           []Tee      `json:"tees"`
	HolesData      []HoleData `json:"holes_data"`
}

type Tee struct {
	TeeKey       string  `json:"tee_key"`
	TeeName      string  `json:"tee_name"`
	TeeColor     string  `json:"tee_color"`
	Gender       string  `json:"gender"`
	CourseRating float64 `json:"course_rating"`
	Slope        int     `json:"slope"`
	Par          int     `json:"par"`
	Yardage      int     `json:"yardage"`
}

type HoleData struct {
	Number          int            `json:"number"`
	Par             int            `json:"par"`
	HandicapIndex   int            `json:"handicap_index"`
	Yardages        map[string]int `json:"yardages"`
	TeeCoords       any            `json:"tee_coords"`
	Green           GreenInfo      `json:"green"`
	GreenPolygon    any            `json:"green_polygon"`
	FairwayPolygon  any            `json:"fairway_polygon"`
	GreenDepthYards *int           `json:"green_depth_yards"`
	GreenWidthYards *int           `json:"green_width_yards"`
	LandingZone     any            `json:"landing_zone"`
	Dogleg          any            `json:"dogleg"`
	Elevation       any            `json:"elevation"`
	PlaysLikeYards  any            `json:"plays_like_yards"`
	Hazards         []any          `json:"hazards"`
}

type GreenInfo struct {
	Center any `json:"center"`
	Front  any `json:"front"`
	Back   any `json:"back"`
}

type CourseSearchResult struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CourseName string  `json:"course_name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	State      *string `json:"state"`
	City       *string `json:"city"`
	Type       *string `json:"type"`
	Par        *int    `json:"par"`
	Phone      any     `json:"phone"`
	Website    any     `json:"website"`
}

type CourseSearchResponse struct {
	Courses []CourseSearchResult `json:"courses"`
	Total   int                  `json:"total"`
}
