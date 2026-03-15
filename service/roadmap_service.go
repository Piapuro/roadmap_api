package service

import (
	"context"
	"database/sql"
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
		if errors.Is(err, sql.ErrNoRows) {
			return response.MVPSuggestionResponse{}, ErrRequirementNotFound
		}
		return response.MVPSuggestionResponse{}, fmt.Errorf("get requirement: %w", err)
	}

	prompt := buildMVPPrompt(req, features)

	raw, err := s.aiAdapter.Generate(ctx, prompt)
	if err != nil {
		return response.MVPSuggestionResponse{}, fmt.Errorf("%w: %w", ErrMVPSuggestionFailed, err)
	}

	return parseMVPResponse(raw)
}

func buildMVPPrompt(req query.Requirement, features []query.RequirementFeature) string {
	featureList := make([]string, 0, len(features))
	for _, f := range features {
		featureList = append(featureList, "- "+f.FeatureName)
	}
	featuresText := strings.Join(featureList, "\n")
	if featuresText == "" {
		featuresText = "（機能リストなし）"
	}

	freeText := "（未入力）"
	if req.FreeText.Valid && req.FreeText.String != "" {
		freeText = req.FreeText.String
	}

	difficultyLabelMap := map[int16]string{1: "初級", 2: "中級", 3: "上級"}
	difficultyLabel, ok := difficultyLabelMap[req.DifficultyLevel]
	if !ok {
		difficultyLabel = fmt.Sprintf("レベル%d", req.DifficultyLevel)
	}

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
		featuresText,
	)
}

// parseMVPResponse はGeminiの応答テキストからJSONオブジェクトを抽出してパースする。
// コードブロック (```json...```) や余分なテキストが含まれていても動作する。
func parseMVPResponse(raw string) (response.MVPSuggestionResponse, error) {
	// 最初の { から最後の } までを抽出してパース
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end < start {
		return response.MVPSuggestionResponse{}, fmt.Errorf("%w: no JSON object found in AI response", ErrMVPSuggestionFailed)
	}
	raw = raw[start : end+1]

	var result response.MVPSuggestionResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return response.MVPSuggestionResponse{}, fmt.Errorf("%w: invalid json from AI: %v", ErrMVPSuggestionFailed, err)
	}
	return result, nil
}
