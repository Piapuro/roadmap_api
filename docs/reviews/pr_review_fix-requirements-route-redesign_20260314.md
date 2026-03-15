# PR レビューレポート

**PR/ブランチ**: `fix/requirements-route-redesign` → `dev`
**PR番号**: #71
**レビュー日時**: 2026-03-14
**変更規模**: +221 / -88 / 12ファイル

---

## 🎯 変更の概要

`POST /requirements`（チームIDをボディに含む設計）を `POST /teams/:id/requirements` に移動し、`TeamScopeAuth.RequireMember()` ミドルウェアによるチームメンバーシップ自動検証を追加した。合わせて `sql/queries/teams.sql` に不足していた 7 クエリを追加して sqlc を再生成し、これまでビルドが通らなかった `adapter/team_adapter.go` と `middleware/team_scope_auth.go` を修正した。

**変更種別**:
- [x] バグ修正 (Bug Fix) — 500エラー原因の根本対応（FK + チームスコープ）
- [x] 新機能 (Feature) — `GET /teams/:id/requirements` 追加
- [x] リファクタリング — NotFound/Locked の区別、sqlcクエリ補完

---

## ✅ マージ判定

> **NEEDS DISCUSSION**

`POST/GET /teams/:id/requirements` のチームスコープ保護は正しく追加された。しかし **既存の `GET/PUT/POST /requirements/:id` にはチームスコープ検証がなく**、JWT を持つ任意のユーザーが他チームの要件定義を参照・更新できる状態が残る。この仕様を意図的に許容するかどうかをチームで決定してからマージすることを推奨する。

---

## 📁 変更ファイル一覧

| ファイル | 変更種別 | +行 | -行 | 懸念度 |
|---------|---------|-----|-----|--------|
| `sql/queries/teams.sql` | Modified | +43 | 0 | 🟡 要確認 |
| `query/teams.sql.go` | Modified | +51 | -64 | 🟢 自動生成 |
| `query/requirements.sql.go` | Modified | +65 | -51 | 🟢 自動生成 |
| `service/requirement_service.go` | Modified | +48 | -20 | 🟡 要確認 |
| `controller/requirement_controller.go` | Modified | +43 | -2 | 🟡 要確認 |
| `router/team_router.go` | Modified | +4 | -1 | 🟢 問題なし |
| `router/requirement_router.go` | Modified | 0 | -1 | 🟢 問題なし |
| `adapter/requirement_adapter.go` | Modified | +9 | 0 | 🟢 問題なし |
| `requests/requirement_request.go` | Modified | +5 | -6 | 🟢 問題なし |
| `dicontainer/dicontainer.go` | Modified | +1 | -1 | 🟢 問題なし |
| `documents/db-design.md` | Modified | +17 | -1 | 🟢 問題なし |

---

## 🔍 詳細レビュー

### service/requirement_service.go

#### 変更の意図
`CreateRequirement` のシグネチャを `teamID` をパスパラメータから受け取る形に変更。`UpdateRequirement` / `LockRequirement` で `ErrNoRows` が「存在しない」と「ロック済み」のどちらかを区別できなかった問題（C-2）を、事前 SELECT で解決。

#### 指摘事項

**[🟡 W-1] `GetTeamRequirements` が features を取得しない** (`service/requirement_service.go:51`)

```go
// 現在のコード（featuresが常に空）
result = append(result, toRequirementResponse(r, nil))
```

> 一覧APIとして features を返さない設計は許容できるが、レスポンスの `features` フィールドが常に `[]` になることが Swagger や API ドキュメントに明記されていない。フロントエンドが features を一覧で使う場合に誤解を招く。

```go
// 推奨: ドキュメントに注記を追加 or features を含む専用レスポンス型を定義
// @Description 一覧取得のため features は含まれません。詳細は GET /requirements/:id を使用してください
```

---

**[🟡 W-2] `UpdateRequirement` / `LockRequirement` の二重クエリ（TOCTOU）** (`service/requirement_service.go:65-90`)

```go
// SELECT してからUPDATE までの間に他リクエストが状態を変更できる
existing, _, err := s.requirementAdapter.GetRequirement(ctx, id)
// ...
if existing.Status == "locked" { return ErrRequirementLocked }
// ← ここで別リクエストが lock する可能性（微小）
updated, features, err := s.requirementAdapter.UpdateRequirement(ctx, id, ...)
```

> 低トラフィックなら実害はほぼないが、厳密にはSELECT〜UPDATEの間にレースコンディションが存在する。将来的に `UPDATE ... WHERE status='draft' RETURNING *` の結果が0行だった場合を先にSELECTで「存在確認」する形にすれば解決する（現状の実装でも運用上は十分許容範囲）。

---

**[🟡 W-3] `IssueInviteToken` サービスが `IsTeamOwner` を二重チェック** (`service/team_service.go:36-42`)

```go
// サービス内でオーナーチェック
isOwner, err := s.teamAdapter.IsTeamOwner(ctx, userID, teamID)
```

> `router/team_router.go` の `ts.RequireOwner()` ミドルウェアで既に検証済みのため、サービス内のチェックは冗長。このPRの変更ではないが、同様のパターンが `GetTeamMembers` にも存在するため、将来のリファクタリング候補として記録しておく。

---

### controller/requirement_controller.go

#### 変更の意図
`CreateRequirement` でパスパラメータ `:id` から `teamID` を取得するよう変更。`ListRequirements` ハンドラ追加。`UpdateRequirement` / `SubmitRequirement` に 404 レスポンスを追加。

#### 指摘事項

**[🔴 C-1] `GET/PUT/POST /requirements/:id` にチームスコープ検証なし** (`router/requirement_router.go:8-14`)

```go
// 現在の requirement_router.go — JWT認証のみ、チーム所属チェックなし
func RegisterRequirementRoutes(e *echo.Echo, c *controller.RequirementController, m *middleware.SupabaseAuth) {
    g := e.Group("/requirements", m.Verify)
    g.GET("/:id", c.GetRequirement)      // 他チームの要件定義も取得可能
    g.PUT("/:id", c.UpdateRequirement)   // 他チームの要件定義も更新可能
    g.POST("/:id/submit", c.SubmitRequirement) // 他チームの要件定義もロック可能
}
```

> JWT を持つ任意のユーザーが要件定義 ID を知っていれば他チームのデータを参照・変更できる。意図的な設計（IDが推測困難なUUIDだから許容）なのか、それとも TeamScopeAuth を追加すべきなのかを明確化する必要がある。
>
> **推奨修正**: `RequirementController.GetRequirement` 内で `requirements.team_id` を取得し、コンテキストの `team_id` と照合するか、`/requirements/:id` を廃止して `/teams/:teamId/requirements/:id` に統合する。

---

### sql/queries/teams.sql

#### 変更の意図
`adapter/team_adapter.go` と `middleware/team_scope_auth.go` が参照していた未定義クエリ（`IssueInviteToken`, `GetTeamByInviteToken`, `IsTeamOwner`, `IsTeamMember`, `JoinTeamAsMember`, `ListTeamMembers`, `GetUserTeamRoleID`）をすべて追加。

#### 指摘事項

**[🟡 W-4] `JoinTeamAsMember` でマジックナンバー使用** (`sql/queries/teams.sql:52`)

```sql
-- name: JoinTeamAsMember :exec
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)  -- ← 1 = TEAM_MEMBER（マジックナンバー）
ON CONFLICT (user_id, team_id) DO NOTHING;
```

> `1` が `TEAM_MEMBER` の ID であることがコードを読むだけではわからない。コメントを追加するか、`(SELECT id FROM team_roles WHERE name = 'TEAM_MEMBER')` を使う。

```sql
-- 推奨
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)  -- 1 = TEAM_MEMBER (team_roles.id)
ON CONFLICT (user_id, team_id) DO NOTHING;
```

---

### requests/requirement_request.go

#### 変更の意図
`TeamID` フィールド削除（パスパラメータから取得）、`DifficultyLevel` のバリデーション上限を `max=5→max=3`（DB CHECK制約と一致）、`ProductType` に `oneof` バリデーション追加。

> バリデーション修正はすべて正しく、DB制約との整合性が取れた。問題なし。

---

## ⚠️ 影響範囲

**このPRの変更が影響する箇所**:
- `dicontainer/dicontainer.go` — `RegisterTeamRoutes` のシグネチャ変更（`requirementController` 追加）
- `router/team_router.go` — `RequirementController` 依存を追加
- クライアント（フロントエンド）— `POST /requirements` が廃止され `POST /teams/:id/requirements` に変更 → **破壊的変更**

**破壊的変更 (Breaking Change)**: **あり**

| 変更前 | 変更後 |
|--------|--------|
| `POST /requirements` (body に `team_id`) | `POST /teams/:id/requirements` (パスパラメータに `team_id`) |
| `GET /requirements/:id` のみ（一覧なし） | `GET /teams/:id/requirements`（一覧）が追加 |

---

## 🧪 テスト確認

| テスト項目 | 状態 |
|-----------|------|
| 既存テストがパスするか | ❓ テストファイルが存在しない |
| 新機能のテストが追加されているか | ❌ 不足 |
| チームスコープ認証のテスト | ❌ 不足 |
| NotFound/Locked の区別テスト | ❌ 不足 |

**追加すべきテストケース**:
```go
// CreateRequirement: チーム非メンバーが 403 を受け取ること
// UpdateRequirement: 存在しないIDで 404、ロック済みで 409 を受け取ること
// LockRequirement: 同上
// ListRequirements: 空チームで [] を返すこと
```

---

## 💬 レビューコメント（コピペ用）

**全体コメント**:
```
レビューしました。

1件のCritical、3件のWarningがあります。

Criticalの C-1（GET/PUT/POST /requirements/:id にチームスコープ検証なし）について
仕様として許容するかどうかをチームで決めてからマージをお願いします。

詳細は docs/reviews/pr_review_fix-requirements-route-redesign_20260314.md を参照してください。
```

**インラインコメント候補**:
- `router/requirement_router.go:11`: `GET /:id` に TeamScopeAuth を追加する必要があります。要件定義の `team_id` を取得してリクエストユーザーの所属チームと照合してください。
- `sql/queries/teams.sql:52`: `1` は `TEAM_MEMBER` の ID です。コメントを追加してください：`-- 1 = TEAM_MEMBER (team_roles.id)`
- `service/requirement_service.go:51`: `GetTeamRequirements` の一覧では features が空になります。Swagger のコメントに明記してください。

---

## ✅ チェックリスト

- [ ] **C-1** `GET/PUT/POST /requirements/:id` のチームスコープ対応方針が決定している
- [ ] `GET /teams/:id/requirements` の features 省略がドキュメントに記載されている
- [ ] Breaking Change（`POST /requirements` 廃止）がフロントエンドチームに共有されている
- [ ] テストが追加・更新されている（または今回スコープ外であることが合意されている）
- [ ] migration `000004_add_requirements` が本番DB（Supabase）に適用済みであることを確認した

---
*Generated by Claude Code / pr-review skill*
