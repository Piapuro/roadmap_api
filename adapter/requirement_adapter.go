package adapter

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Piapuro/roadmap_api/query"
	"github.com/google/uuid"
)

type RequirementAdapter struct {
	q  *query.Queries
	db *sql.DB
}

func NewRequirementAdapter(q *query.Queries, db *sql.DB) *RequirementAdapter {
	return &RequirementAdapter{q: q, db: db}
}

type RequirementInput struct {
	TeamID          uuid.UUID
	ProductType     string
	DifficultyLevel int16
	FreeText        *string
	SupplementURL   *string
	CreatedBy       uuid.UUID
	Features        []string
}

// CreateRequirement は requirements と requirement_features をトランザクションで一括登録する。
func (a *RequirementAdapter) CreateRequirement(ctx context.Context, in RequirementInput) (query.Requirement, []query.RequirementFeature, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := a.q.WithTx(tx)

	req, err := qtx.CreateRequirement(ctx, query.CreateRequirementParams{
		TeamID:          in.TeamID,
		ProductType:     in.ProductType,
		DifficultyLevel: in.DifficultyLevel,
		FreeText:        nullString(in.FreeText),
		SupplementUrl:   nullString(in.SupplementURL),
		CreatedBy:       in.CreatedBy,
	})
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("create requirement: %w", err)
	}

	features, err := insertFeatures(ctx, qtx, req.ID, in.Features)
	if err != nil {
		return query.Requirement{}, nil, err
	}

	if err := tx.Commit(); err != nil {
		return query.Requirement{}, nil, fmt.Errorf("commit: %w", err)
	}
	return req, features, nil
}

// HasConfirmedRoadmap は requirement が属するチームに confirmed なロードマップがあるか返す。
func (a *RequirementAdapter) HasConfirmedRoadmap(ctx context.Context, teamID uuid.UUID) (bool, error) {
	return a.q.HasConfirmedRoadmapForTeam(ctx, teamID)
}

// GetRequirement は requirement と関連する features を取得する。
func (a *RequirementAdapter) GetRequirement(ctx context.Context, id uuid.UUID) (query.Requirement, []query.RequirementFeature, error) {
	req, err := a.q.GetRequirementByID(ctx, id)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("get requirement: %w", err)
	}
	features, err := a.q.ListRequirementFeatures(ctx, id)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("list features: %w", err)
	}
	return req, features, nil
}

// ListRequirementsByTeamID はチームに属する要件定義一覧を返す。
func (a *RequirementAdapter) ListRequirementsByTeamID(ctx context.Context, teamID uuid.UUID) ([]query.Requirement, error) {
	reqs, err := a.q.ListRequirementsByTeamID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("list requirements by team: %w", err)
	}
	return reqs, nil
}

type RequirementUpdateInput struct {
	ProductType     string
	DifficultyLevel int16
	FreeText        *string
	SupplementURL   *string
	Features        []string
}

// UpdateRequirement は draft 状態の requirement を更新し、features を差し替える。
func (a *RequirementAdapter) UpdateRequirement(ctx context.Context, id uuid.UUID, in RequirementUpdateInput) (query.Requirement, []query.RequirementFeature, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := a.q.WithTx(tx)

	req, err := qtx.UpdateRequirement(ctx, query.UpdateRequirementParams{
		ID:              id,
		ProductType:     in.ProductType,
		DifficultyLevel: in.DifficultyLevel,
		FreeText:        nullString(in.FreeText),
		SupplementUrl:   nullString(in.SupplementURL),
	})
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("update requirement: %w", err)
	}

	if err := qtx.DeleteRequirementFeatures(ctx, id); err != nil {
		return query.Requirement{}, nil, fmt.Errorf("delete features: %w", err)
	}

	features, err := insertFeatures(ctx, qtx, id, in.Features)
	if err != nil {
		return query.Requirement{}, nil, err
	}

	if err := tx.Commit(); err != nil {
		return query.Requirement{}, nil, fmt.Errorf("commit: %w", err)
	}
	return req, features, nil
}

// LockRequirement は draft → locked へステータスを遷移させる。
func (a *RequirementAdapter) LockRequirement(ctx context.Context, id uuid.UUID) (query.Requirement, []query.RequirementFeature, error) {
	req, err := a.q.LockRequirement(ctx, id)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("lock requirement: %w", err)
	}
	features, err := a.q.ListRequirementFeatures(ctx, id)
	if err != nil {
		return query.Requirement{}, nil, fmt.Errorf("list features: %w", err)
	}
	return req, features, nil
}

func insertFeatures(ctx context.Context, q *query.Queries, reqID uuid.UUID, names []string) ([]query.RequirementFeature, error) {
	features := make([]query.RequirementFeature, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		f, err := q.CreateRequirementFeature(ctx, query.CreateRequirementFeatureParams{
			RequirementID: reqID,
			FeatureName:   name,
		})
		if err != nil {
			return nil, fmt.Errorf("create feature %q: %w", name, err)
		}
		features = append(features, f)
	}
	return features, nil
}

func nullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
