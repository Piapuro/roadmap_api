package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/requests"
	"github.com/Piapuro/roadmap_api/response"
	"github.com/google/uuid"
)

var (
	ErrRequirementNotFound = errors.New("requirement not found")
	ErrRequirementLocked   = errors.New("requirement is already locked")
)

type RequirementService struct {
	requirementAdapter *adapter.RequirementAdapter
}

func NewRequirementService(requirementAdapter *adapter.RequirementAdapter) *RequirementService {
	return &RequirementService{requirementAdapter: requirementAdapter}
}

func (s *RequirementService) CreateRequirement(ctx context.Context, userID uuid.UUID, req requests.CreateRequirementRequest) (response.RequirementResponse, error) {
	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		return response.RequirementResponse{}, fmt.Errorf("invalid team_id: %w", err)
	}

	req2, features, err := s.requirementAdapter.CreateRequirement(ctx, adapter.RequirementInput{
		TeamID:          teamID,
		ProductType:     req.ProductType,
		DifficultyLevel: int16(req.DifficultyLevel),
		FreeText:        nilIfEmpty(req.FreeText),
		SupplementURL:   nilIfEmpty(req.SupplementURL),
		CreatedBy:       userID,
		Features:        req.Features,
	})
	if err != nil {
		return response.RequirementResponse{}, err
	}
	return toRequirementResponse(req2, features), nil
}

func (s *RequirementService) GetRequirement(ctx context.Context, id uuid.UUID) (response.RequirementResponse, error) {
	req, features, err := s.requirementAdapter.GetRequirement(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.RequirementResponse{}, ErrRequirementNotFound
		}
		return response.RequirementResponse{}, err
	}
	return toRequirementResponse(req, features), nil
}

func (s *RequirementService) UpdateRequirement(ctx context.Context, id uuid.UUID, req requests.UpdateRequirementRequest) (response.RequirementResponse, error) {
	updated, features, err := s.requirementAdapter.UpdateRequirement(ctx, id, adapter.RequirementUpdateInput{
		ProductType:     req.ProductType,
		DifficultyLevel: int16(req.DifficultyLevel),
		FreeText:        nilIfEmpty(req.FreeText),
		SupplementURL:   nilIfEmpty(req.SupplementURL),
		Features:        req.Features,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.RequirementResponse{}, ErrRequirementLocked
		}
		return response.RequirementResponse{}, err
	}
	return toRequirementResponse(updated, features), nil
}

func (s *RequirementService) LockRequirement(ctx context.Context, id uuid.UUID) (response.RequirementResponse, error) {
	locked, features, err := s.requirementAdapter.LockRequirement(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.RequirementResponse{}, ErrRequirementLocked
		}
		return response.RequirementResponse{}, err
	}
	return toRequirementResponse(locked, features), nil
}

func toRequirementResponse(r query.Requirement, features []query.RequirementFeature) response.RequirementResponse {
	featureNames := make([]string, 0, len(features))
	for _, f := range features {
		featureNames = append(featureNames, f.FeatureName)
	}
	resp := response.RequirementResponse{
		ID:              r.ID.String(),
		TeamID:          r.TeamID.String(),
		ProductType:     r.ProductType,
		DifficultyLevel: int(r.DifficultyLevel),
		Status:          r.Status,
		Features:        featureNames,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if r.FreeText.Valid {
		resp.FreeText = r.FreeText.String
	}
	if r.SupplementUrl.Valid {
		resp.SupplementURL = r.SupplementUrl.String
	}
	return resp
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
