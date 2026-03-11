# 03 重要な実装パターンと概念

## 1. sqlc によるタイプセーフな DB アクセス

### 仕組み

このプロジェクトでは **ORM を使いません**。代わりに `sqlc` を使って、SQL クエリから Go のコードを自動生成します。

```
sql/queries/users.sql  →  sqlc generate  →  query/users.sql.go
sql/schema.sql         →                 →  query/models.go
```

### 手順

```bash
# 1. sql/queries/*.sql に SQL を書く
-- name: CreateUserSkill :one
INSERT INTO user_skills (user_id, skill_name, experience_years, is_learning_goal)
VALUES ($1, $2, $3, $4)
RETURNING *;

# 2. sqlc generate を実行
sqlc generate

# 3. 自動生成された関数を使う（手書き不要）
skill, err := q.CreateUserSkill(ctx, query.CreateUserSkillParams{
    UserID:    userID,
    SkillName: "Go",
    ...
})
```

### なぜ sqlc を使うのか

| 比較 | ORM (GORM等) | sqlc |
|------|------------|------|
| SQL の可読性 | 低い（メソッドチェーン） | 高い（生 SQL） |
| パフォーマンス | 遅くなりがち | 生 SQL と同等 |
| 型安全性 | 弱い | 強い（コンパイル時エラー） |
| SQL インジェクション対策 | 自動 | 自動（プリペアドステートメント） |
| 複雑なクエリ | 難しい | そのまま書ける |

---

## 2. トランザクション管理パターン

`adapter/user_adapter.go` の `UpsertSkills` が模範的な実装例です。

```go
func (a *UserAdapter) UpsertSkills(ctx context.Context, ...) error {
    // パターン1: BeginTx + defer Rollback
    tx, err := a.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback()  // ← 重要: Commit 後の Rollback は no-op なので安全

    // パターン2: WithTx でトランザクション付き Queries を作る
    qtx := a.q.WithTx(tx)  // query.Queries の db フィールドを tx に差し替える

    // ... 複数のクエリを実行 ...

    // パターン3: 最後に Commit（失敗したら defer の Rollback が動く）
    return tx.Commit()
}
```

**`defer tx.Rollback()` が安全な理由**:
- `tx.Commit()` が成功した後に `Rollback()` を呼んでも、PostgreSQL はすでに確定済みとして無視します
- `return fmt.Errorf(...)` や `panic` でどこから抜けても必ずロールバックされます

---

## 3. エラーハンドリングの2段階

Go では `error` は値として扱います。このプロジェクトでは2つのパターンを使います。

### パターンA: ラップしてそのまま上位に渡す（Adapter 層）

```go
// adapter/user_adapter.go
user, err := a.q.GetUserByID(ctx, userID)
if err != nil {
    return query.User{}, nil, fmt.Errorf("get user: %w", err)
    //                                   ↑ %w でラップすることで errors.Is が使える
}
```

### パターンB: 種別を判定して HTTP ステータスを決める（Controller 層）

```go
// controller/user_controller.go
user, skills, err := c.userService.GetMySkills(...)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        // ユーザーが見つからない → 404
        return ctx.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
    }
    // それ以外 → 500
    return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
```

> **`errors.Is` が使えるのはなぜか？**: `fmt.Errorf("...: %w", err)` で包んでも、`errors.Is` はラップの中を再帰的に探してくれるからです。

---

## 4. Null 値の扱い（`sql.NullString`, `sql.NullUUID`）

PostgreSQL の NULL は Go の `sql.Null*` 型で表現します。

```go
// query/models.go（sqlc 生成）
type User struct {
    ID           uuid.UUID
    Email        string
    Name         string
    PasswordHash sql.NullString  // VARCHAR(255) NULL → NullString
    AvatarUrl    sql.NullString  // VARCHAR(500) NULL → NullString
    Bio          sql.NullString  // VARCHAR(200) NULL → NullString
    SkillLevel   string          // VARCHAR(20) NOT NULL → string
}
```

### 読み取り時

```go
bio := ""
if user.Bio.Valid {      // NULL でなければ
    bio = user.Bio.String
}
```

### 書き込み時

```go
sql.NullString{
    String: bio,
    Valid:  bio != "",  // 空文字の場合は NULL として保存
}
```

---

## 5. UUID を使った主キー管理

全テーブルの主キーは `UUID` です。`github.com/google/uuid` で扱います。

```go
// 文字列 → uuid.UUID（Controller で JWT の sub を変換）
userID, err := uuid.Parse(str)

// uuid.UUID → 文字列（レスポンスで返す時）
s.ID.String()
```

> **UUID を使う理由**: 連番 ID と違い、DB の外で ID を事前生成できます。分散環境でも衝突しません。

---

## 6. バリデーションタグ

`requests` パッケージの構造体フィールドに書くタグで、`ctx.Validate()` 時に自動検証されます。

```go
// requests/skill_request.go
type UpsertSkillsRequest struct {
    SkillLevel string       `json:"skill_level" validate:"required,oneof=beginner intermediate advanced"`
    Bio        string       `json:"bio"         validate:"max=200"`
    Skills     []SkillInput `json:"skills"      validate:"required"`
}

type SkillInput struct {
    SkillName       string   `json:"skill_name" validate:"required,max=30"`
    ExperienceYears *float64 `json:"experience_years"`  // ポインタ = オプション
    IsLearningGoal  bool     `json:"is_learning_goal"`
}
```

| タグ | 意味 |
|------|------|
| `required` | ゼロ値は不可（空文字・nil・0・false は NG） |
| `max=30` | 文字列の最大長 30 |
| `oneof=a b c` | a か b か c のいずれか |

> **`*float64` vs `float64`**: ポインタ型にするとフィールドが省略可能（nil = 未入力）になります。`float64` だと 0.0 と「未入力」の区別ができません。

---

## 7. Echo のコンテキスト（`echo.Context`）

Echo の `ctx` はリクエスト情報のコンテナです。よく使うメソッドをまとめます。

```go
// リクエスト
ctx.Bind(&req)          // JSON Body → 構造体
ctx.Validate(&req)      // バリデーション実行
ctx.Param("id")         // パスパラメータ（/teams/:id の :id）
ctx.QueryParam("page")  // クエリパラメータ（?page=1）

// コンテキスト値（middleware が書き込んだ値を読む）
ctx.Get("user_id")      // interface{} で返る → 型アサーションが必要
ctx.Set("key", value)   // 値を書き込む

// レスポンス
ctx.JSON(http.StatusOK, data)       // JSON 返却
ctx.NoContent(http.StatusNoContent) // 204

// リクエストの Go context（DB 操作に渡す）
ctx.Request().Context()
```

---

## 8. DB スキーマの設計思想

### CHECK 制約による列挙値の強制

```sql
-- skill_level は "beginner", "intermediate", "advanced" のみ許可
skill_level VARCHAR(20) NOT NULL DEFAULT 'beginner'
    CHECK (skill_level IN ('beginner', 'intermediate', 'advanced'))
```

アプリ側のバリデーションと DB 側の制約をダブルで保護しています。

### ON DELETE CASCADE

```sql
-- ユーザーを削除したら、そのユーザーのスキルも自動削除
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
```

### global_roles / team_roles の数値レベル

```sql
INSERT INTO global_roles (id, name, level) VALUES
    (1, 'GUEST', 10),
    (2, 'LOGIN_USER', 20),
    (3, 'SYSTEM_ADMIN', 99)
```

`level` を数値にすることで「level >= 20 なら許可」という比較が簡単に書けます。

---

## 9. Docker 構成のポイント

### 本番 Dockerfile（マルチステージビルド）

```dockerfile
# Stage 1: ビルド（golang イメージ）
FROM golang:1.24-alpine AS builder
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s' \  # デバッグ情報を削除してバイナリを小さく
    -o /app/api main.go

# Stage 2: 実行（distroless = シェルなし最小イメージ）
FROM gcr.io/distroless/static-debian12
USER nonroot:nonroot  # root で動かさない（セキュリティ）
```

**distroless を使う理由**: シェルもパッケージマネージャもないため、コンテナへの侵入時にできることが極めて限られます。

### 開発 docker-compose の依存関係

```yaml
api:
  depends_on:
    db:
      condition: service_healthy  # db が healthy になるまで api を起動しない
```

`service_healthy` は DB の `healthcheck` が PASS するまで待ちます。単純な `depends_on: db` より確実です。

---

## 10. 命名規則まとめ

| 対象 | 規則 | 例 |
|------|------|-----|
| ファイル | スネークケース | `user_controller.go` |
| 型・構造体 | アッパーキャメル | `UserController`, `UpsertSkillsRequest` |
| メソッド | アッパーキャメル | `GetMySkills`, `UpsertMySkills` |
| 変数 | ローワーキャメル | `userID`, `expYears` |
| SQL クエリ名 | アッパーキャメル | `-- name: CreateUserSkill :one` |
| JSON フィールド | スネークケース | `"skill_name"`, `"is_learning_goal"` |
| DB テーブル/列 | スネークケース | `user_skills`, `experience_years` |
| ルートパス | ケバブケース | `/users/me/skills` |

---

## つまずきやすいポイント

### Q: `query/` フォルダのファイルを手書きで編集してしまった

**A**: sqlc の生成ファイルです。次回 `sqlc generate` を実行すると上書きされます。SQL を変更したい場合は `sql/queries/*.sql` を編集してから `sqlc generate` を実行します。

### Q: バリデーションが効かない

**A**: `ctx.Bind()` だけではバリデーションは実行されません。必ず `ctx.Validate(&req)` もセットで呼びます。また `dicontainer.go` で `e.Validator = utils.NewValidator()` を設定していないと `ctx.Validate()` は panic します。

### Q: `sql.ErrNoRows` が 500 になる

**A**: DB が「レコードなし」を返した時の `sql.ErrNoRows` は 500 ではなく 404 です。`errors.Is(err, sql.ErrNoRows)` で判定して適切なステータスを返します。

### Q: `ctx.Get("user_id")` が nil になる

**A**: 認証ミドルウェアを通っていないルートです。ルーターで `m.Verify` を設定しているグループか確認します。また、型アサーション `raw.(string)` が失敗している可能性もあります。
