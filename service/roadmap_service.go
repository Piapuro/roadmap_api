package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Piapuro/roadmap_api/adapter"
	"github.com/Piapuro/roadmap_api/query"
	"github.com/Piapuro/roadmap_api/response"
	"github.com/google/uuid"
)

var ErrMVPSuggestionFailed = errors.New("MVP suggestion failed")

type RoadmapService struct {
	aiAdapter          *adapter.AIAdapter
	requirementAdapter *adapter.RequirementAdapter
}

func NewRoadmapService(aiAdapter *adapter.AIAdapter, requirementAdapter *adapter.RequirementAdapter) *RoadmapService {
	return &RoadmapService{
		aiAdapter:          aiAdapter,
		requirementAdapter: requirementAdapter,
	}
}

// SuggestMVP は要件定義データをもとに AI が MVP 機能リストを提案する。
func (s *RoadmapService) SuggestMVP(ctx context.Context, requirementID uuid.UUID) (response.MVPSuggestionResponse, error) {
	req, features, err := s.requirementAdapter.GetRequirement(ctx, requirementID)
	if err != nil {
		return response.MVPSuggestionResponse{}, fmt.Errorf("get requirement: %w", err)
	}

	prompt := buildMVPPrompt(req, features)

	raw, err := s.aiAdapter.Generate(ctx, prompt)
	if err != nil {
		return response.MVPSuggestionResponse{}, fmt.Errorf("%w: %v", ErrMVPSuggestionFailed, err)
	}

	return parseMVPResponse(raw)
}

func buildMVPPrompt(req query.Requirement, features []query.RequirementFeature) string {
	featureList := make([]string, 0, len(features))
	for _, f := range features {
		featureList = append(featureList, "- "+f.FeatureName)
	}

	freeText := ""
	if req.FreeText.Valid {
		freeText = req.FreeText.String
	}

	difficultyLabel := map[int16]string{1: "初級", 2: "中級", 3: "上級"}[req.DifficultyLevel]

	return fmt.Sprintf(`あなたはソフトウェアプロジェクトのMVP（最小実行可能製品）設計の専門家です。
以下の要件定義からMVP（最初に開発すべき最小限の機能）を提案してください。

プロダクトタイプ: %s
難易度レベル: %s
概要: %s
機能一覧:
%s

必ず以下のJSON形式のみで回答してください（説明文は不要）:
{
  "mvp_features": ["機能名1", "機能名2"],
  "reasoning": "選定理由の説明"
}`,
		req.ProductType,
		difficultyLabel,
		freeText,
		strings.Join(featureList, "\n"),
	)
}

// parseMVPResponse はGeminiの応答テキストからJSONを抽出してパースする。
func parseMVPResponse(raw string) (response.MVPSuggestionResponse, error) {
	// コードブロック (```json ... ```) を除去
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result response.MVPSuggestionResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return response.MVPSuggestionResponse{}, fmt.Errorf("%w: invalid json from AI: %v", ErrMVPSuggestionFailed, err)
	}
	return result, nil
}
