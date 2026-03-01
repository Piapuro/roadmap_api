package adapter

import "github.com/your-name/roadmap/api/query"

type RequirementAdapter struct {
	q *query.Queries
}

func NewRequirementAdapter(q *query.Queries) *RequirementAdapter {
	return &RequirementAdapter{q: q}
}
