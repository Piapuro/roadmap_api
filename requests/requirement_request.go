package requests

type CreateRequirementRequest struct {
	ProductType     string   `json:"product_type"     validate:"required,oneof=web app game ai" example:"web"`
	DifficultyLevel int      `json:"difficulty_level" validate:"required,min=1,max=3"            example:"2"`
	FreeText        string   `json:"free_text"                                                   example:"ECサイトを作りたい"`
	SupplementURL   string   `json:"supplement_url"                                              example:"https://example.com/spec"`
	Features        []string `json:"features"                                                    example:"ログイン機能,商品一覧,カート機能"`
}

type UpdateRequirementRequest struct {
	ProductType     string   `json:"product_type"     validate:"omitempty,oneof=web app game ai" example:"app"`
	DifficultyLevel int      `json:"difficulty_level" validate:"omitempty,min=1,max=3"           example:"3"`
	FreeText        string   `json:"free_text"                                                   example:"モバイルアプリに変更"`
	SupplementURL   string   `json:"supplement_url"                                              example:"https://example.com/spec-v2"`
	Features        []string `json:"features"                                                    example:"プッシュ通知,オフライン対応"`
}
