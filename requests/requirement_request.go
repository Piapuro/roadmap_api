package requests

type CreateRequirementRequest struct {
	TeamID          string   `json:"team_id"          validate:"required,uuid"        example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductType     string   `json:"product_type"     validate:"required"             example:"web"`
	DifficultyLevel int      `json:"difficulty_level" validate:"required,min=1,max=5" example:"2"`
	FreeText        string   `json:"free_text"                                        example:"ECサイトを作りたい"`
	SupplementURL   string   `json:"supplement_url"                                   example:"https://example.com/spec"`
	Features        []string `json:"features"                                         example:"ログイン機能,商品一覧,カート機能"`
}

type UpdateRequirementRequest struct {
	ProductType     string   `json:"product_type"                                    example:"app"`
	DifficultyLevel int      `json:"difficulty_level" validate:"omitempty,min=1,max=5" example:"3"`
	FreeText        string   `json:"free_text"                                       example:"モバイルアプリに変更"`
	SupplementURL   string   `json:"supplement_url"                                  example:"https://example.com/spec-v2"`
	Features        []string `json:"features"                                        example:"プッシュ通知,オフライン対応"`
}
