package course

type Course struct {
	ID         string
	Name       string
	ClubName   string
	City       string
	State      string
	Lat        float64
	Lng        float64
	Type       string
	Par        int
	Holes      int
	Yardage    int
	Timezone   string
	YearBuilt  int
	Phone      string
	Website    string
	Address    string
	PostalCode string

	Tees      []Tee
	HolesData []Hole
}

type Tee struct {
	Key          string
	Name         string
	Color        string
	Gender       string
	CourseRating float64
	Slope        int
	Par          int
	Yardage      int
}

type Hole struct {
	Number        int
	Par           int
	HandicapIndex int
	Yardages      map[string]int
}
