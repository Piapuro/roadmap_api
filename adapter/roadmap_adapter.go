package adapter

import "github.com/your-name/roadmap/api/query"

type RoadmapAdapter struct {
	q *query.Queries
}

func NewRoadmapAdapter(q *query.Queries) *RoadmapAdapter {
	return &RoadmapAdapter{q: q}
}
