# セキュリティレビューレポート

**レビュー日時**: 2026-03-12
**対象**: 認証・認可周り全体（middleware/, controller/auth_controller.go, service/auth_service.go, router/*, dicontainer/）
**スキャン項目数**: 13

---

## 🚨 リスクサマリー

| 深刻度 | 件数 | 対応優先度 |
|--------|------|-----------|
| 🔴 Critical（即座に対応） | 1 | 今すぐ |
| 🟠 High（今週中）         | 4 | 今週中 |
| 🟡 Medium（今月中）       | 5 | 今月中 |
| 🟢 Informational（推奨）  | 4 | 任意 |

**総合リスクレベル**: **CRITICAL**

---

## 🔴 Critical

### [SEC-C1] 本番シークレットが `.env.local` に平文で存在

**カテゴリ**: Secret Leak
**ファイル**: `.env.local`
**CVSS スコア**: ~9.8/10

**問題**:
`.env.local` に実際の本番プロジェクト `qedzamfpuoyqfatayzqp.supabase.co` のシークレットが平文で記載されている。

```
SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_SERVICE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
SUPABASE_JWT_SECRET=oz5GUkL6opZd...
WEBHOOK_SECRET=eb35e016...
GEMINI_API_KEY=AIzaSyBes...
```

**リスク**: `SERVICE_KEY` は管理者権限相当。漏洩した場合、全データへの無制限アクセスが可能になる。`JWT_SECRET` 漏洩は任意の有効トークン偽造を可能にする。

**修正方法**:
1. 上記シークレットをすべて**今すぐローテート**（Supabase ダッシュボード → Settings → API）
2. `git log` で過去コミットへの混入がないか確認
3. `git-secrets` 等を導入してコミット時に自動検出

**参考**: [OWASP A02: Cryptographic Failures]

---

## 🟠 High

### [SEC-H1] JWT の audience（aud）クレーム検証が未実装

**カテゴリ**: Authentication Bypass
**ファイル**: `middleware/supabase_auth.go:128`

**問題のコード**:
```go
// ❌ WithAudience が未指定
token, err := jwt.ParseWithClaims(tokenStr, claims, keyfunc, jwt.WithExpirationRequired())
```

**リスク**: 別サービス向けに発行されたトークン（同一 issuer を持つ）が受け入れられる可能性がある。

**修正方法**:
```go
// ✅
token, err := jwt.ParseWithClaims(tokenStr, claims, keyfunc,
    jwt.WithExpirationRequired(),
    jwt.WithAudience("authenticated"),
)
```

---

### [SEC-H2] 内部エラーの詳細がHTTPレスポンスに露出

**カテゴリ**: Information Disclosure
**ファイル**: `controller/user_controller.go:115, 156` / `controller/team_controller.go:55`

**問題のコード**:
```go
// ❌ DBのテーブル名・カラム名・内部ロジックが露出する
return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
```

**修正方法**:
```go
// ✅ ログに詳細を残し、レスポンスは固定文字列のみ
log.Printf("internal error: %v", err)
return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
```

---

### [SEC-H3] CORS ミドルウェアが Echo に未登録

**カテゴリ**: Security Misconfiguration
**ファイル**: `dicontainer/dicontainer.go:87`

**問題**: `config.CORSAllowOrigins` が定義・読み込まれているにもかかわらず、Echo に一切適用されていない。

**修正方法**:
```go
// ✅ dicontainer.go の e.Use() に追加
e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
    AllowOrigins:     strings.Split(cfg.CORSAllowOrigins, ","),
    AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders:     []string{echo.HeaderAuthorization, echo.HeaderContentType},
    AllowCredentials: true,
}))
```

---

### [SEC-H4] ログイン・サインアップにレート制限なし

**カテゴリ**: Brute Force / Credential Stuffing
**ファイル**: `router/auth_router.go:11`

**問題**: `/auth/signup`, `/auth/login` にレート制限がない。`/webhooks` には `RateLimiter` が適用済みなのに認証エンドポイントが無防備。

**修正方法**:
```go
// ✅ auth グループに RateLimiter を追加
g := e.Group("/auth")
g.Use(echoMiddleware.RateLimiter(echoMiddleware.NewRateLimiterMemoryStore(20)))
// 本番では Redis バックエンドに変更（docker-compose に redis サービスあり）
```

---

## 🟡 Medium

### [SEC-M1] 認可（RBAC）が全エンドポイントで未適用

**カテゴリ**: Missing Authorization
**ファイル**: `router/*.go` 全体

**問題**: `RBAC` 関数は実装済みだが、どのルートにも適用されていない。ロールをDBから取得してコンテキストにセットするミドルウェアも存在しない。
**現状**: 「認証済みであれば全操作が可能」な状態。他ユーザーのリソース操作や不正な削除が理論上可能。

**対応**: RBACを実際のルートに適用する実装を優先的に進める。

---

### [SEC-M2] Supabase エラーボディをそのままログに出力

**カテゴリ**: Sensitive Data Exposure
**ファイル**: `service/auth_service.go:105`

**問題のコード**:
```go
// ❌ メールアドレスが含まれる可能性があるボディをそのままログ出力
log.Printf("[AuthService] Supabase login error: status=%d body=%v", resp.StatusCode, errBody)
```

**修正方法**:
```go
// ✅ error_code と msg のみログ出力
log.Printf("[AuthService] Supabase login error: status=%d code=%v", resp.StatusCode, errBody["error_code"])
```

---

### [SEC-M3] issuer mismatch ログにログインジェクションの余地

**カテゴリ**: Log Injection
**ファイル**: `middleware/supabase_auth.go:155`

**問題**: 攻撃者が送信したトークンの `issuer` 値がそのままログに出力される。改行文字を含む値で偽ログエントリを作成可能。

**修正方法**:
```go
// ✅ 制御文字を除去してからログ出力
safeIssuer := strings.Map(func(r rune) rune {
    if r == '\n' || r == '\r' || r < 32 { return -1 }
    return r
}, claims.Issuer)
log.Printf("[SupabaseAuth] issuer mismatch: got=%q want=%q", safeIssuer, m.issuer)
```

---

### [SEC-M4] JWKS キャッシュが起動時の1回取得のみ

**カテゴリ**: Availability / Key Rotation
**ファイル**: `middleware/supabase_auth.go:47`

**問題**: Supabase が鍵をローテートすると、再起動するまで ES256 トークンがすべて拒否される。

**修正方法**: `kid` が見つからない場合に `fetchJWKS()` を再試行するフォールバック、またはバックグラウンド goroutine で定期更新（例: 1時間ごと）を追加する。

---

### [SEC-M5] Webhook 署名検証が平文比較（HMAC 未使用）

**カテゴリ**: Insufficient Verification
**ファイル**: `controller/webhook_controller.go:41`

**問題**: `subtle.ConstantTimeCompare` でタイミング攻撃は防いでいるが、Supabase 公式 Webhook の署名方式（HMAC-SHA256）と異なる。

**修正方法**: `x-supabase-signature` ヘッダーを使った HMAC-SHA256 署名検証に移行する。

---

## 🟢 Informational

### [SEC-I1] ✅ alg:none 攻撃は適切にブロック済み

`middleware/supabase_auth.go` の型スイッチにより HS256/ES256 以外を拒否。`jwt_test.go` にテストケースあり。対応不要。

### [SEC-I2] ✅ トークン有効期限検証は適切に実施

`jwt.WithExpirationRequired()` で `exp` クレームを必須化・検証済み。対応不要。

### [SEC-I3] Logout コントローラーに冗長なトークン取り出し処理

**ファイル**: `controller/auth_controller.go:92`
`m.Verify` ミドルウェアで手前に認証済みのため実害なし。コンテキストの `user_id` を使う方式に統一することを推奨。

### [SEC-I4] GET /skills が認証なし公開

静的なマスターデータのため設計上は問題ない可能性が高いが、`webhook_router.go` のように意図をコメントで明示することを推奨。

---

## ✅ セキュリティチェックリスト

### 認証・認可
- [x] ほぼすべての状態変更エンドポイントに JWT 認証が設定されている
- [x] alg:none 攻撃が防御されている
- [x] トークン有効期限が検証されている
- [x] issuer 検証が実装されている
- [ ] audience（aud）検証が実装されていない → **SEC-H1**
- [ ] RBAC がいずれのルートにも適用されていない → **SEC-M1**

### データ保護
- [ ] 本番シークレットが .env.local に平文で存在 → **SEC-C1（即時対応）**
- [x] .env.local は .gitignore に含まれている（ただし漏洩リスクあり）
- [ ] 500 エラー時に内部情報がレスポンスに含まれる → **SEC-H2**

### 通信・インフラ
- [ ] CORS ミドルウェアが未適用 → **SEC-H3**
- [x] DB クエリはすべてパラメータ化クエリ（sqlc 使用）
- [x] Webhook シークレット検証にタイミング攻撃対策あり

### レート制限
- [ ] 認証エンドポイントにレート制限なし → **SEC-H4**
- [x] Webhook エンドポイントに RateLimiter 適用済み

---

## 📋 対応優先度まとめ

| 優先度 | ID | 内容 | 工数目安 |
|--------|----|------|---------|
| 🔴 今すぐ | SEC-C1 | 全シークレットのローテート | 30分 |
| 🟠 今週 | SEC-H1 | audience 検証追加 | 15分 |
| 🟠 今週 | SEC-H2 | err.Error() をレスポンスから除去 | 30分 |
| 🟠 今週 | SEC-H3 | CORS ミドルウェア登録 | 30分 |
| 🟠 今週 | SEC-H4 | 認証エンドポイントにレート制限追加 | 1時間 |
| 🟡 今月 | SEC-M1 | RBAC をルートに適用 | 1〜2日 |
| 🟡 今月 | SEC-M2 | ログのメールアドレス除去 | 30分 |
| 🟡 今月 | SEC-M3 | ログインジェクション対策 | 30分 |
| 🟡 今月 | SEC-M4 | JWKS 定期更新 | 2時間 |
| 🟡 今月 | SEC-M5 | Webhook HMAC 署名検証 | 2時間 |

---

*Generated by Claude Code / security-review skill — 2026-03-12*
