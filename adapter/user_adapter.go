package adapter

import "github.com/your-name/roadmap/api/query"

type UserAdapter struct {
	q *query.Queries
}

func NewUserAdapter(q *query.Queries) *UserAdapter {
	return &UserAdapter{q: q}
}
