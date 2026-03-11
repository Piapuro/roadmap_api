package adapter

import "github.com/Piapuro/roadmap_api/query"

type RequirementAdapter struct {
	q *query.Queries
}

func NewRequirementAdapter(q *query.Queries) *RequirementAdapter {
	return &RequirementAdapter{q: q}
}
