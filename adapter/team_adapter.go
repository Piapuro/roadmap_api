package adapter

import "github.com/your-name/roadmap/api/query"

type TeamAdapter struct {
	q *query.Queries
}

func NewTeamAdapter(q *query.Queries) *TeamAdapter {
	return &TeamAdapter{q: q}
}
