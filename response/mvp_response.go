package response

// MVPSuggestionResponse は AI が提案する MVP 機能リストと理由を返す。
type MVPSuggestionResponse struct {
	MVPFeatures []string `json:"mvp_features" example:"[\"ユーザー登録\",\"ログイン\"]"`
	Reasoning   string   `json:"reasoning"    example:"ユーザー認証は全機能の基盤となるため最優先です"`
}
