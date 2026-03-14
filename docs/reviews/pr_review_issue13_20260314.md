# PR レビューレポート

**PR**: #70 feat: 要件定義DBスキーマ・CRUD API実装 (#13)
**ブランチ**: `feature/issue13-requirements-schema`
**レビュー日時**: 2026-03-14
**変更規模**: +688 / -43 / 13ファイル

---

## 🎯 変更の概要

要件定義（requirements / requirement_features）のDBマイグレーション・SQLクエリ・全レイヤーのCRUD APIを実装したPR。
ローカル向け migration 000004 と Supabase 向けマイグレーションの両方を追加し、既存のスタブ実装を完全に置き換えた。

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

致命的なバグではないが、**バリデーション不整合（DifficultyLevel max=5 vs DB CHECK max=3）** と **NotFound/Locked のエラー区別不能** の2点はユーザーへの誤ったエラーレスポンスにつながるため、修正が必要。
それ以外の実装品質（トランザクション設計、sqlcクエリ、レイヤー分離）は良好。

---

## 📁 変更ファイル一覧

| ファイル | 変更種別 | +行 | -行 | 懸念度 |
|---------|---------|-----|-----|--------|
| `requests/requirement_request.go` | 既存 (未変更) | - | - | 🔴 バリデーション不整合 |
| `sql/queries/requirements.sql` | 新規 | +45 | - | 🟡 未使用クエリあり |
| `sql/migrations/000004_*.sql` | 新規 | +15 | - | 🟢 問題なし |
| `supabase/migrations/20260314000004_add_requirements.sql` | 新規 | +30 | - | 🟢 問題なし |
| `adapter/requirement_adapter.go` | 大幅追加 | +154 | -3 | 🟡 軽微 |
| `service/requirement_service.go` | 大幅追加 | +109 | - | 🔴 エラー区別不能 |
| `controller/requirement_controller.go` | 大幅追加 | +88 | -14 | 🟡 権限チェック不足 |
| `query/requirements.sql.go` | 新規(生成) | +237 | - | 🟢 問題なし |
| `query/teams.sql.go` | sqlc再生成 | +35 | - | 🟢 問題なし |
| `dicontainer/dicontainer.go` | 軽微修正 | +1 | -1 | 🟢 問題なし |

---

## 🔍 詳細レビュー

### [🔴 C-1] DifficultyLevel のバリデーション上限がDBと不整合

**ファイル**: `requests/requirement_request.go:6,14`

```go
// ❌ 現状: max=5 を許可しているが DB は 1〜3 しか受け付けない
DifficultyLevel int `json:"difficulty_level" validate:"required,min=1,max=5"`
```

```sql
-- DB CHECK 制約
difficulty_level SMALLINT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 3)
```

```go
// ✅ 修正: max を 3 に合わせる
DifficultyLevel int `json:"difficulty_level" validate:"required,min=1,max=3"`
```

> `difficulty_level=4` や `5` を送ると DB の CHECK 制約違反で 500 Internal Server Error になる。ユーザーには 400 Bad Request を返すべき。`UpdateRequirementRequest` も同様に修正が必要。

---

### [🔴 C-2] LockRequirement / UpdateRequirement で NotFound と Locked のエラーが区別できない

**ファイル**: `service/requirement_service.go:70-82`

```go
// ❌ 現状: UpdateRequirement/LockRequirement の ErrNoRows は
//   「IDが存在しない」と「status=draft でない（locked）」の両方で発生する
if errors.Is(err, sql.ErrNoRows) {
    return response.RequirementResponse{}, ErrRequirementLocked
}
```

SQLクエリに `AND status = 'draft'` があるため、存在しないIDの場合も `ErrNoRows` になる。
存在しないIDに対して 409 Conflict が返ってしまう。

```go
// ✅ 修正案: 先に存在確認を行う
func (s *RequirementService) UpdateRequirement(ctx context.Context, id uuid.UUID, req requests.UpdateRequirementRequest) (response.RequirementResponse, error) {
    // 先に requirement を取得して存在・状態を確認
    existing, _, err := s.requirementAdapter.GetRequirement(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return response.RequirementResponse{}, ErrRequirementNotFound
        }
        return response.RequirementResponse{}, err
    }
    if existing.Status == "locked" {
        return response.RequirementResponse{}, ErrRequirementLocked
    }
    // 更新処理へ...
}
```

---

### [🟡 W-1] `GetRequirementByTeamID` クエリが実装されているが未使用

**ファイル**: `sql/queries/requirements.sql:13-17`

```sql
-- name: GetRequirementByTeamID :one
SELECT * FROM requirements
WHERE team_id = $1
ORDER BY created_at DESC
LIMIT 1;
```

生成された `query/requirements.sql.go` に関数は存在するが、adapter/service/controller のどこからも呼ばれていない。
将来の実装のために残す場合は、コメントに意図を書くか、チームの要件定義取得 API (`GET /teams/:id/requirement`) として実装を追加することを推奨。

---

### [🟡 W-2] `requirement_features.is_required` が常に `true` で固定

**ファイル**: `adapter/requirement_adapter.go:131`

```go
// ❌ is_required を指定していないため DB デフォルト (true) 固定になる
f, err := q.CreateRequirementFeature(ctx, query.CreateRequirementFeatureParams{
    RequirementID: reqID,
    FeatureName:   name,
})
```

DBスキーマと `RequirementFeature` モデルには `is_required` フィールドがあるが、APIから制御できない。
現時点でユーザーが `is_required=false` を指定する手段がない。

```go
// ✅ 将来の拡張を考慮する場合: 構造体でフラグを受け取る
type FeatureInput struct {
    Name       string
    IsRequired bool
}
```

> MVP では `is_required=true` 固定でも許容範囲。ただし `requests.CreateRequirementRequest.Features` が `[]string` のままでは将来的に拡張しにくいため、Sprint 3 完了前に設計を確認することを推奨。

---

### [🟡 W-3] `CreateRequirement` にチームメンバー権限チェックがない

**ファイル**: `controller/requirement_controller.go:37-59`

要件定義は `team_id` に紐づくが、そのチームのメンバーかどうかを確認していない。
任意のユーザーが任意のチームに要件定義を作成できてしまう。

```go
// ❌ team_id が body で指定されるため TeamScopeAuth MW が使えない
// service.CreateRequirement 内でチームメンバーチェックが必要
```

> Issue #12 の `TeamScopeAuth` はパスパラメータ `:id` 前提のため適用できない。
> `service.CreateRequirement` 内で `IsTeamMember` を呼んでチェックするか、
> ルーティングを `/teams/:id/requirements` に変更することを検討。

---

### [🟢 Good] トランザクション設計が適切

```go
// adapter/requirement_adapter.go
tx, err := a.db.BeginTx(ctx, nil)
defer func() { _ = tx.Rollback() }()
// ... requirements INSERT + features INSERT
tx.Commit()
```

`defer Rollback()` パターンを使用しており、途中でエラーが発生した場合にロールバックが確実に実行される。`UserAdapter` と同じパターンで一貫性もある。

---

### [🟢 Good] SQL での状態制御が安全

```sql
-- LockRequirement: draft でない場合は ErrNoRows で失敗するため二重ロックを防止
UPDATE requirements
SET status = 'locked', updated_at = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING *;
```

アプリケーション層でステータスをチェックするより安全。ただし C-2 の問題（NotFound との区別）が伴う。

---

## ⚠️ 影響範囲

**このPRの変更が影響する箇所**:
- `dicontainer/dicontainer.go` — `NewRequirementAdapter` に `db` が追加（破壊的変更なし・内部変更のみ）
- `query/teams.sql.go` — sqlc 再生成で `AssignTeamOwner` クエリが追加（機能追加のみ）
- `router/requirement_router.go` — 変更なし（既存ルート定義はそのまま機能する）

**破壊的変更 (Breaking Change)**: なし

---

## 🧪 テスト確認

| テスト項目 | 状態 |
|-----------|------|
| 既存テストがパスするか | ❓ 未確認（テストなし） |
| 新機能のテストが追加されているか | ❌ 不足 |
| DifficultyLevel=4 送信時の挙動 | ❌ 500になる（要修正） |
| 存在しないIDに PUT した場合 | ❌ 409になる（要修正） |

**追加すべきテストケース**:
```go
// service_test.go
func TestUpdateRequirement_NotFound(t *testing.T) {
    // 存在しないIDへのUpdate → ErrRequirementNotFound
}
func TestUpdateRequirement_Locked(t *testing.T) {
    // locked状態のUpdate → ErrRequirementLocked
}
func TestCreateRequirement_InvalidDifficulty(t *testing.T) {
    // DifficultyLevel=4 → バリデーションエラー
}
```

---

## 💬 レビューコメント（コピペ用）

**全体コメント**:
```
レビューしました。

実装の全体的な品質（トランザクション設計、レイヤー分離、sqlcクエリ）は良好です。
ただし以下2点の修正をお願いします。

🔴 C-1: DifficultyLevel のバリデーション上限が max=5 になっていますが、
         DB CHECK 制約は 1〜3 です。requests/requirement_request.go の max を 3 に修正してください。

🔴 C-2: UpdateRequirement / LockRequirement で存在しない ID に対して
         409 Conflict が返ってしまいます。先に GetRequirement で存在確認を行い、
         NotFound と Locked を区別してください。

修正後に再レビューします。
```

**インラインコメント候補**:
- `requests/requirement_request.go:6`: `max=5` → `max=3` に修正してください（DB の CHECK 制約に合わせる）
- `service/requirement_service.go:70`: `ErrNoRows` は「IDなし」と「locked」の両方で発生します。事前に `GetRequirement` で確認してください
- `sql/queries/requirements.sql:13`: `GetRequirementByTeamID` はどこから使われますか？未使用であればコメントで意図を書いてください

---

## ✅ チェックリスト

- [ ] **C-1** `DifficultyLevel` バリデーション max を 3 に修正
- [ ] **C-2** UpdateRequirement / LockRequirement で NotFound / Locked を区別
- [ ] W-1: `GetRequirementByTeamID` の使用箇所を実装 or コメント追記
- [ ] テストが追加されている（最低限 C-1, C-2 のケース）
- [ ] セルフレビュー済み

---
*Generated by Claude Code / pr-review skill*
