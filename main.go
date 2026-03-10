// Package main is the entry point of the Roadmap API server.
package main

import (
	"log"

	_ "github.com/your-name/roadmap/api/docs"
	"github.com/your-name/roadmap/api/dicontainer"
)

// @title           Roadmap API
// @version         1.0
// @description     チーム開発ロードマップ生成API
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     "Bearer {Supabase JWT token}" の形式で入力してください
func main() {
	container, err := dicontainer.New()
	if err != nil {
		log.Fatalf("failed to initialize container: %v", err)
	}

	if err := container.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
