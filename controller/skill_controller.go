package controller

import (
	"net/http"

	"github.com/Piapuro/roadmap_api/response"
	"github.com/labstack/echo/v4"
)

type SkillController struct{}

func NewSkillController() *SkillController {
	return &SkillController{}
}

var skillTags = []response.SkillTagResponse{
	// frontend
	{Name: "HTML/CSS", Category: "frontend"},
	{Name: "JavaScript", Category: "frontend"},
	{Name: "TypeScript", Category: "frontend"},
	{Name: "React", Category: "frontend"},
	{Name: "Vue.js", Category: "frontend"},
	{Name: "Next.js", Category: "frontend"},
	// backend
	{Name: "Go", Category: "backend"},
	{Name: "Python", Category: "backend"},
	{Name: "Node.js", Category: "backend"},
	{Name: "Java", Category: "backend"},
	{Name: "Ruby", Category: "backend"},
	{Name: "Rust", Category: "backend"},
	// database
	{Name: "PostgreSQL", Category: "database"},
	{Name: "MySQL", Category: "database"},
	{Name: "Redis", Category: "database"},
	{Name: "MongoDB", Category: "database"},
	{Name: "SQLite", Category: "database"},
	// infra
	{Name: "Docker", Category: "infra"},
	{Name: "Kubernetes", Category: "infra"},
	{Name: "AWS", Category: "infra"},
	{Name: "GCP", Category: "infra"},
	{Name: "Terraform", Category: "infra"},
	// ai
	{Name: "Machine Learning", Category: "ai"},
	{Name: "Deep Learning", Category: "ai"},
	{Name: "LLM", Category: "ai"},
	{Name: "Computer Vision", Category: "ai"},
	{Name: "NLP", Category: "ai"},
	// mobile
	{Name: "Swift", Category: "mobile"},
	{Name: "Kotlin", Category: "mobile"},
	{Name: "React Native", Category: "mobile"},
	{Name: "Flutter", Category: "mobile"},
}

func (c *SkillController) ListSkillTags(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, skillTags)
}
