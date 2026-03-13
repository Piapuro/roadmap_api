# PR レビューレポート

**PR/ブランチ**: issue10 → main ([PR #66](https://github.com/Piapuro/roadmap_api/pull/66))
**対象コミット**: `0529db5` feat: 招待リンク発行・参加API実装 (#10)
**レビュー日時**: 2026-03-12
**変更規模**: +350行 / 8ファイル

---

## 🎯 変更の概要

Issue #10「招待リンク発行・参加API」の実装。チームオーナーが招待トークン（有効期限7日）を発行し、そのトークンを使って別ユーザーがチームにTEAM_MEMBERとして参加できるAPIを追加した。

**変更種別**:
- [x] 新機能 (Feature)
- [ ] バグ修正 (Bug Fix)
- [ ] リファクタリング
- [ ] パフォーマンス改善
- [ ] テスト追加
- [ ] ドキュメント

---

## ✅ マージ判定

> **REQUEST CHANGES**

コア機能の実装は正確で、セキュリティ設計（`crypto/rand`、適切なエラー分類）も良好。ただし **2点のクリティカルな問題**（`sql.ErrNoRows`の検出漏れ・既存トークン上書き時の通知なし）と、テストの不在がある。修正後にマージを推奨。

---

## 📁 変更ファイル一覧

| ファイル | 変更種別 | +行 | 懸念度 |
|---------|---------|-----|--------|
| `service/team_service.go` | Modified | +72 | 🔴 要確認 |
| `adapter/team_adapter.go` | Modified | +46 | 🟢 問題なし |
| `controller/team_controller.go` | Modified | +88 | 🟡 軽微 |
| `query/teams.sql.go` | Modified | +98 | 🟡 軽微 |
| `router/team_router.go` | Modified | +2 | 🟢 問題なし |
| `requests/team_request.go` | Modified | +4 | 🟢 問題なし |
| `response/team_response.go` | Modified | +13 | 🟢 問題なし |
| `sql/queries/teams.sql` | Modified | +27 | 🟢 問題なし |

---

## 🔍 詳細レビュー

### service/team_service.go

#### 変更の意図
招待トークン発行と参加処理のビジネスロジックを実装。オーナー確認 → トークン生成 → DB保存、および トークン検索 → 期限確認 → 重複チェック → 参加の2フローを担う。

#### 指摘事項

**[🔴 C-1] `sql.ErrNoRows` がアダプタのラップにより検出されない** (`service/team_service.go:64`)

```go
// ❌ 現状: adapter が fmt.Errorf("get team by invite token: %w", err) でラップするため
// errors.Is(err, sql.ErrNoRows) は真になる（%w でチェーン保持される）が、
// adapter が返すエラーが "not found" を示すかどうか確認が必要。
// GetTeamByInviteToken で存在しないトークンを渡すと sql.ErrNoRows が返る。
// ただし adapter.GetTeamByInviteToken がラップしているので、
// errors.Is は %w チェーンで辿れるため実際には動作する。
// → 問題は adapter 層で "not found" を別エラーとして明示しない一貫性の欠如。
team, err := s.teamAdapter.GetTeamByInviteToken(ctx, token)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {  // %w でラップされていても Is() は通るが...
        return response.JoinTeamResponse{}, ErrInviteTokenNotFound
    }
```

```go
// ✅ 推奨: adapter 層でドメインエラーに変換する（他のadapterとの一貫性のため）
// adapter/team_adapter.go:
func (a *TeamAdapter) GetTeamByInviteToken(ctx context.Context, token string) (query.Team, error) {
    team, err := a.q.GetTeamByInviteToken(ctx, token)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return query.Team{}, fmt.Errorf("get team by invite token: %w", ErrTeamNotFound)
        }
        return query.Team{}, fmt.Errorf("get team by invite token: %w", err)
    }
    return team, nil
}
```

> 現状の `errors.Is(err, sql.ErrNoRows)` は `%w` により機能するが、adapter の責務としてドメインエラーへの変換を行うべき。他の adapter メソッドとの一貫性の問題。

---

**[🔴 C-2] 既存の有効なトークンを上書きする際に通知・確認なし** (`service/team_service.go:32`)

```go
// ❌ 現状: 既存の有効トークンがあっても無条件に上書き
func (s *TeamService) IssueInviteToken(...) {
    // ...
    team, err := s.teamAdapter.IssueInviteToken(ctx, teamID, token, expiresAt)
```

```go
// ✅ 推奨: 既存トークンが有効な場合はそのまま返す（またはUIで警告）
func (s *TeamService) IssueInviteToken(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (response.InviteTokenResponse, error) {
    // ... owner check ...

    // 既存の有効トークンがあればそれを返す
    existing, err := s.teamAdapter.GetTeamByID(ctx, teamID)
    if err == nil && existing.InviteToken.Valid && existing.InviteTokenExpiresAt.Valid {
        if time.Now().UTC().Before(existing.InviteTokenExpiresAt.Time) {
            return response.InviteTokenResponse{
                TeamID:    existing.ID.String(),
                Token:     existing.InviteToken.String,
                InviteURL: "/teams/join?token=" + existing.InviteToken.String,
                ExpiresAt: existing.InviteTokenExpiresAt.Time,
            }, nil
        }
    }
    // 期限切れまたは未発行の場合のみ新規生成
    // ...
}
```

> 有効な招待リンクを共有済みの場合に再発行すると古いリンクが無効化される。誤操作で参加者が困るリスク。少なくとも「既存トークンがある」ことをレスポンスに含めるか、上書き前に有効期限を確認すべき。

---

**[🟡 W-1] チームが存在しない場合のハンドリングなし** (`controller/team_controller.go:140`)

```go
// ❌ 存在しないteamIDを渡すと DB がレコードなしのエラーを返すが、500になる
resp, err := c.teamService.IssueInviteToken(ctx.Request().Context(), userID, teamID)
if err != nil {
    if errors.Is(err, service.ErrNotTeamOwner) {
        return ctx.JSON(http.StatusForbidden, ...)
    }
    return ctx.JSON(http.StatusInternalServerError, ...) // ← 404 であるべき
}
```

```go
// ✅ ErrTeamNotFound を追加して 404 を返す
if errors.Is(err, service.ErrTeamNotFound) {
    return ctx.JSON(http.StatusNotFound, map[string]string{"error": "team not found"})
}
```

---

**[🟡 W-2] `JoinTeamAsMember` の `ON CONFLICT DO NOTHING` により重複参加が無音で成功する可能性** (`query/teams.sql.go:253`)

```sql
-- 現状: CONFLICT 時に何も返さない
INSERT INTO user_team_roles (user_id, team_id, team_role_id)
VALUES ($1, $2, 1)
ON CONFLICT DO NOTHING;
```

> `IsTeamMember` チェック後に `JoinTeamAsMember` を呼ぶため、TOCTOU (Time-of-Check-Time-of-Use) の競合状態が理論上発生しうる。ただし今の規模では実用上問題なし。将来的には `ON CONFLICT DO NOTHING RETURNING` でチェックを不要にできる。

---

### controller/team_controller.go

#### 指摘事項

**[🟢 I-1] `IssueInviteToken` と `JoinTeam` の userID 取得ロジックが重複**

```go
// 両ハンドラで同じ3行が繰り返されている
userIDStr, ok := ctx.Get(middleware.ContextKeyUserID).(string)
if !ok || userIDStr == "" { ... }
userID, err := uuid.Parse(userIDStr)
```

> 現時点ではDRY原則の軽微な違反。既存コード（`CreateTeam`）も同様のパターンのため、このPRのスコープ外。将来的にヘルパー関数化を検討。

---

### router/team_router.go

#### 変更の意図
`/join` を `/:id` より先に登録することで、Echo の静的ルート優先マッチングを活用。

```go
// ✅ 正しい順序: 静的ルート /join が /:id より先
g.POST("/join", c.JoinTeam)
g.GET("/:id", c.GetTeam)
```

> Echo は静的パスを動的パスより優先するため問題なし。意図的な設計として良い。

---

## ⚠️ 影響範囲

**このPRの変更が影響する箇所**:
- `dicontainer/dicontainer.go` — 変更なし（TeamAdapter/Service/Controller は既存のまま注入）
- `router/team_router.go` — 新ルート追加のみ、既存ルートに影響なし
- `teams` テーブル — `invite_token` / `invite_token_expires_at` カラムは既存スキーマに存在
- `user_team_roles` テーブル — `ON CONFLICT DO NOTHING` で安全に書き込み

**破壊的変更 (Breaking Change)**: なし

---

## 🧪 テスト確認

| テスト項目 | 状態 |
|-----------|------|
| 既存テストがパスするか | ✅ `go build ./...` `go vet ./...` 成功 |
| `IssueInviteToken` のユニットテスト | ❌ 不足 |
| `JoinTeam` のユニットテスト | ❌ 不足 |
| エッジケース（期限切れ・重複・不正トークン） | ❌ 不足 |

**追加すべきテストケース**:

```go
// service/team_service_test.go

func TestIssueInviteToken_NotOwner(t *testing.T) {
    // IsTeamOwner が false を返す場合 ErrNotTeamOwner が返ること
}

func TestIssueInviteToken_Success(t *testing.T) {
    // トークンが64文字の hex であること
    // ExpiresAt が約7日後であること
    // InviteURL が /teams/join?token= で始まること
}

func TestJoinTeam_ExpiredToken(t *testing.T) {
    // InviteTokenExpiresAt が過去の場合 ErrInviteTokenExpired が返ること
}

func TestJoinTeam_NotFound(t *testing.T) {
    // 存在しないトークンで ErrInviteTokenNotFound が返ること
}

func TestJoinTeam_AlreadyMember(t *testing.T) {
    // IsTeamMember が true の場合 ErrAlreadyTeamMember が返ること
}
```

---

## 💬 レビューコメント（コピペ用）

**全体コメント**:
```
レビューしました。

実装方針・セキュリティ設計（crypto/rand 使用、適切なエラー分類）は良好です。
2件のクリティカルな指摘と、テスト不足があります。

Critical:
- C-1: GetTeamByInviteToken の "not found" エラーはアダプタ層で変換推奨（動作はするが設計一貫性の問題）
- C-2: 有効な既存トークンを無条件上書きするため、共有済みリンクが無効化されるリスクあり

Warning:
- W-1: 存在しないteamIDで IssueInviteToken を呼ぶと 500 になる（404 を返すべき）
- W-2: TOCTOU 競合は現在の規模では実用上問題なし（将来対応可）

C-2 の修正（または仕様として許容する場合はコメント追記）と
W-1 の 404 ハンドリング追加をお願いします。
```

**インラインコメント候補**:
- `service/team_service.go:64`: `sql.ErrNoRows` の検出はアダプタ層で行うことを推奨。現状も動作するが、他のアダプタとの一貫性のため。
- `service/team_service.go:32`: 有効な既存トークンがある場合の上書き挙動を明示的にしてください。仕様として許容する場合はコメントを。
- `controller/team_controller.go:140`: `ErrTeamNotFound` を追加して 404 を返せると UX が向上します。

---

## ✅ チェックリスト

- [ ] C-1: adapter での not-found エラー変換（または現状を仕様として明文化）
- [ ] C-2: 有効な既存トークンの上書き挙動を仕様化・またはガード追加
- [ ] W-1: チームが存在しない場合の 404 ハンドリング
- [ ] ユニットテストの追加（少なくともサービス層のエラーケース）
- [ ] 破壊的変更なし ✅
- [ ] `go build` / `go vet` 成功 ✅

---
*Generated by Claude Code / pr-review skill*
