# PR レビューレポート

**PR/ブランチ**: `issue14` → `dev`
**PR番号**: #72
**レビュー日時**: 2026-03-14
**変更規模**: +928 / -304 / 17ファイル

---

## 🎯 変更の概要

Issue #14（[#025] 要件定義保存API）の完全実装。`POST /teams/:id/requirements`（新規作成）と `PUT /requirements/:id`（上書き保存）を中心に、チームスコープ権限チェック・バリデーション・エラー区別・Swagger更新までをカバーする。合わせて `sql/queries/teams.sql` に不足していた7クエリ（`IssueInviteToken`, `IsTeamOwner`, `IsTeamMember` 等）を補完し、これまでビルドが通らなかった `team_adapter` と `team_scope_auth` を修正。

**変更種別**:
- [x] 新機能 (Feature) — POST/GET /teams/:id/requirements 追加
- [x] バグ修正 (Bug Fix) — sqlc クエリ欠落・500エラー・NotFound/Locked混同の修正
- [x] ドキュメント — Swagger・db-design.md・issue-assignment.md 更新

---

## ✅ マージ判定

> **APPROVE（条件付き）**

実装の品質・設計ともに問題なし。指摘事項はすべて軽微（W/I レベル）であり、ブロッカーなし。ただし **C-1（`/requirements/:id` のチームスコープ未検証）** は前回 PR #71 から持ち越された既知の懸念事項であり、チームで仕様合意後に別 Issue として対応することを推奨する。

---

## 📁 変更ファイル一覧

| ファイル | 変更種別 | 懸念度 |
|---------|---------|--------|
| `sql/queries/teams.sql` | +43行 | 🟡 新規クエリ追加（要動作確認） |
| `query/teams.sql.go` | 自動生成 | 🟢 問題なし |
| `query/requirements.sql.go` | 自動生成 | 🟢 問題なし |
| `service/requirement_service.go` | +48行 | 🟡 二重クエリあり（許容範囲） |
| `controller/requirement_controller.go` | +43行 | 🟢 問題なし |
| `router/team_router.go` | +4行 | 🟢 問題なし |
| `router/requirement_router.go` | -1行 | 🟢 問題なし |
| `adapter/requirement_adapter.go` | +9行 | 🟢 問題なし |
| `requests/requirement_request.go` | 修正 | 🟢 問題なし |
| `dicontainer/dicontainer.go` | +1行 | 🟢 問題なし |
| `docs/swagger.yaml` / `docs.go` / `swagger.json` | 更新 | 🟢 問題なし |
| `documents/db-design.md` | +78行 | 🟢 問題なし |
| `documents/issue-assignment.md` | 更新 | 🟢 問題なし |

---

## 🔍 詳細レビュー

### router/team_router.go ✅

#### 変更の意図
`RequirementController` を引数に受け取り、`GET/POST /:id/requirements` を `RequireMember()` 保護下に追加。

```go
// 変更後（正しい設計）
func RegisterTeamRoutes(e *echo.Echo, c *controller.TeamController, rc *controller.RequirementController, ...) {
    g.GET("/:id/requirements", rc.ListRequirements, ts.RequireMember())
    g.POST("/:id/requirements", rc.CreateRequirement, ts.RequireMember())
}
```

> チーム非メンバーは自動的に 403 で弾かれる。設計通り。

---

### requests/requirement_request.go ✅

#### 変更の意図
`team_id` フィールド削除（パスパラメータへ移行）、`DifficultyLevel` の上限をDB制約と合わせて 5→3 に修正、`ProductType` に `oneof` バリデーション追加。

```go
// 変更後（正しい）
ProductType     string `json:"product_type" validate:"required,oneof=web app game ai"`
DifficultyLevel int    `json:"difficulty_level" validate:"required,min=1,max=3"`
```

> DB の `CHECK (product_type IN ('web', 'app', 'game', 'ai'))` および `CHECK (difficulty_level BETWEEN 1 AND 3)` と完全一致。バリデーションによる事前チェックで DB エラーを防止。

---

### service/requirement_service.go

#### 変更の意図
`CreateRequirement` のシグネチャを変更（teamID をパス由来の引数に）。`UpdateRequirement` / `LockRequirement` で ErrNoRows が「存在しない」と「ロック済み」の両方を表していた問題を事前 SELECT で解消。

#### 指摘事項

**[🟡 W-1] UpdateRequirement / LockRequirement の SELECT→UPDATE 間のレースコンディション** (`service/requirement_service.go:65-90`)

```go
// 事前チェック
existing, _, err := s.requirementAdapter.GetRequirement(ctx, id)
if existing.Status == "locked" { return ErrRequirementLocked }

// ↑ と ↓ の間に別リクエストが status を変更する可能性（微小）
updated, features, err := s.requirementAdapter.UpdateRequirement(ctx, id, ...)
```

> 低トラフィックなら実害なし。厳密には `UPDATE ... WHERE id=$1 AND status='draft'` の返り値が 0 行の場合は前の SELECT で既存確認済みなので「ロック済み」と確定できるため、将来的なリファクタリング候補として記録。現状は許容範囲。

---

**[🟡 W-2] GetTeamRequirements が features を空で返す** (`service/requirement_service.go:51`)

```go
result = append(result, toRequirementResponse(r, nil))  // features = []
```

> 一覧 API で features を省略するのは合理的な設計（N+1 防止）だが、Swagger の description に「一覧では features は空。詳細は `GET /requirements/:id` を参照」という注記が現時点でないため、FE 実装者が混乱する可能性がある。

---

### sql/queries/teams.sql 🟡

#### 変更の意図
`adapter/team_adapter.go` と `middleware/team_scope_auth.go` が参照していた未定義クエリ7本を追加。

#### 指摘事項

**[🟡 W-3] JoinTeamAsMember のマジックナンバー** (`sql/queries/teams.sql:52`)

```sql
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)  -- ← 1 の意味が不明
```

```sql
-- 推奨（コメントを追加）
VALUES ($1, $2, 1)  -- 1 = TEAM_MEMBER (team_roles.id)
```

---

### controller/requirement_controller.go ✅

#### 変更の意図
`ListRequirements` ハンドラ追加。`CreateRequirement` でパスパラメータから `teamID` を取得。`UpdateRequirement` / `SubmitRequirement` に 404 レスポンスを追加。

```go
// 404 対応（追加済み）
if errors.Is(err, service.ErrRequirementNotFound) {
    return ctx.JSON(http.StatusNotFound, map[string]string{"error": "requirement not found"})
}
```

> エラーの種類に応じた HTTP ステータス（404 vs 409）が正しく返る。問題なし。

---

### docs/swagger.yaml ✅

Swagger に `/teams/{id}/requirements` の GET・POST が正しく追加されており、`PUT /requirements/{id}` の 404 レスポンスも反映済み。

---

## ⚠️ 影響範囲

**このPRが影響する箇所**:
- `dicontainer/dicontainer.go` — `RegisterTeamRoutes` シグネチャ変更
- フロントエンド — `POST /requirements`（廃止）→ `POST /teams/:id/requirements`（移行必要）

**破壊的変更 (Breaking Change)**: **あり**
- `POST /requirements` エンドポイントが削除された
- `CreateRequirementRequest` から `team_id` フィールドが削除された

**既知の懸念事項（前 PR #71 からの持ち越し）**:

**[🔴 C-1] `GET/PUT/POST /requirements/:id` にチームスコープ検証なし**

```go
// router/requirement_router.go — JWT のみ、チーム所属チェックなし
g.GET("/:id", c.GetRequirement)
g.PUT("/:id", c.UpdateRequirement)
g.POST("/:id/submit", c.SubmitRequirement)
```

> JWT を持つ任意のユーザーが ID を知っていれば他チームのデータを閲覧・更新可能。意図的な設計（UUID はランダムで推測困難）として許容するか、TeamScopeAuth を追加するか、チームで仕様決定が必要。

---

## 🧪 テスト確認

| テスト項目 | 状態 |
|-----------|------|
| 既存テストがパスするか | ❓ テストなし |
| POST /teams/:id/requirements の正常系 | ❌ 未テスト |
| PUT /requirements/:id の 404 / 409 分岐 | ❌ 未テスト |
| チーム非メンバーの 403 | ❌ 未テスト |
| DifficultyLevel=4 で 400 | ❌ 未テスト |
| ProductType 不正値で 400 | ❌ 未テスト |

---

## 💬 レビューコメント（コピペ用）

**全体コメント**:
```
レビューしました。

Criticalなブロッカーはなく、APPROVE です。

ただし以下の点を今後の Issue として積み上げることを推奨します：
1. GET/PUT/POST /requirements/:id のチームスコープ検証（C-1）
2. GET /teams/:id/requirements の Swagger に「features は空」を明記（W-2）

詳細は docs/reviews/pr_review_issue14_20260314.md を参照してください。
```

**インラインコメント候補**:
- `router/requirement_router.go:11`: `GET /:id` は現在チームスコープ外です。将来的に `PUT` も含めて TeamScopeAuth を追加予定であれば Issue を立てておきましょう。
- `service/requirement_service.go:51`: 一覧では features が空になります。`// NOTE: 一覧では features は取得しません。詳細は GetRequirement を使用してください` のコメントを追加すると FE チームへの伝達が楽になります。
- `sql/queries/teams.sql:52`: `VALUES ($1, $2, 1)` の `1` に `-- 1 = TEAM_MEMBER` のコメントを追加してください。

---

## ✅ チェックリスト

- [x] ビルドが通る（`go build ./...` 確認済み）
- [x] Swagger に新エンドポイントが反映されている
- [x] db-design.md・issue-assignment.md が更新されている
- [ ] C-1（`/requirements/:id` チームスコープ）の対応方針が Issue 化されている
- [ ] Breaking Change（`POST /requirements` 廃止）がフロントエンドチームに共有されている
- [ ] migration `000004` が本番 DB（Supabase）に適用済みであること

---
*Generated by Claude Code / pr-review skill*
