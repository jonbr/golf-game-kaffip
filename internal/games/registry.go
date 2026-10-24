package games

import "github.com/go-chi/chi/v5"

// Module is one self-contained game format. Each format registers its own
// routes: nothing outside its folder needs to know how it works.
type Module interface {
	Type() string
	RegisterRoutes(r chi.Router)
}
