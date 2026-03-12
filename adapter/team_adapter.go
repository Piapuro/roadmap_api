package adapter

import "github.com/Piapuro/roadmap_api/query"

type TeamAdapter struct {
	q *query.Queries
}

func NewTeamAdapter(q *query.Queries) *TeamAdapter {
	return &TeamAdapter{q: q}
}
