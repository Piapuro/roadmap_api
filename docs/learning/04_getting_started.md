# 04 開発スタートガイド

## セットアップ手順

### 前提条件

- Go 1.24+
- Docker Desktop
- sqlc（`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`）

### 1. 環境変数を設定する

```bash
cp .env.example .env.local
```

`.env.local` を編集：

```env
# Supabase プロジェクト設定
SUPABASE_URL=https://xxxx.supabase.co
SUPABASE_ANON_KEY=eyJ...
SUPABASE_SERVICE_KEY=eyJ...
SUPABASE_JWT_SECRET=your-jwt-secret

# ローカル Docker の PostgreSQL（docker-compose のデフォルト）
DATABASE_URL=postgres://dev_user:dev_password@localhost:5432/roadmap_dev?sslmode=disable

# Google Gemini API
GEMINI_API_KEY=AIza...

# CORS
CORS_ALLOW_ORIGINS=http://localhost:3000
```

### 2. Docker でローカル環境を起動する

```bash
# 初回（ビルド + マイグレーション実行）
docker compose up --build

# 2回目以降
docker compose up
```

起動後:
- API サーバー: `http://localhost:8080`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

Air（ホットリロード）が有効なので、Go ファイルを保存すると自動でビルド・再起動されます。

### 3. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health
# → {"status":"ok"}

# スキルタグ一覧（認証不要）
curl http://localhost:8080/skills
```

---

## 新しいエンドポイントを追加する手順

チームの一覧を返す `GET /teams` を実装する例で説明します。

### Step 1: SQL クエリを書く

```sql
-- sql/queries/teams.sql に追記
-- name: ListTeamsByUserID :many
SELECT t.* FROM teams t
INNER JOIN user_team_roles utr ON t.id = utr.team_id
WHERE utr.user_id = $1
ORDER BY t.created_at DESC;
```

### Step 2: sqlc generate を実行する

```bash
sqlc generate
```

`query/teams.sql.go` に `ListTeamsByUserID` 関数が生成されます。

### Step 3: Adapter に関数を追加する

```go
// adapter/team_adapter.go
func (a *TeamAdapter) ListByUserID(ctx context.Context, userID uuid.UUID) ([]query.Team, error) {
    teams, err := a.q.ListTeamsByUserID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("list teams: %w", err)
    }
    return teams, nil
}
```

### Step 4: Service にビジネスロジックを追加する

```go
// service/team_service.go
func (s *TeamService) GetMyTeams(ctx context.Context, userID uuid.UUID) ([]query.Team, error) {
    return s.teamAdapter.ListByUserID(ctx, userID)
}
```

### Step 5: Response 型を定義する

```go
// response/team_response.go
type TeamResponse struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Goal      string `json:"goal"`
    Level     string `json:"level"`
}
```

### Step 6: Controller にハンドラーを追加する

```go
// controller/team_controller.go
func (c *TeamController) GetTeams(ctx echo.Context) error {
    userID, err := parseUserID(ctx)
    if err != nil {
        return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
    }

    teams, err := c.teamService.GetMyTeams(ctx.Request().Context(), userID)
    if err != nil {
        return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    res := make([]response.TeamResponse, len(teams))
    for i, t := range teams {
        res[i] = response.TeamResponse{
            ID:    t.ID.String(),
            Name:  t.Name,
            Goal:  t.Goal,
            Level: t.Level,
        }
    }
    return ctx.JSON(http.StatusOK, res)
}
```

### Step 7: ルートに登録する

```go
// router/team_router.go（既存の g.GET("/", c.GetTeams) を確認・実装）
g.GET("", c.GetTeams)  // GET /teams
```

### Step 8: ビルドして確認する

```bash
go build ./...
```

エラーがなければ完成です。

---

## よくあるタスクのパターン

### パターン A: バリデーションが必要な POST エンドポイント

```go
func (c *XxxController) Create(ctx echo.Context) error {
    var req requests.CreateXxxRequest
    if err := ctx.Bind(&req); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    if err := ctx.Validate(&req); err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    // ... service 呼び出し
    return ctx.JSON(http.StatusCreated, res)
}
```

### パターン B: パスパラメータから ID を取得する

```go
func (c *XxxController) GetByID(ctx echo.Context) error {
    id, err := uuid.Parse(ctx.Param("id"))
    if err != nil {
        return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
    }
    // ... service 呼び出し
}
```

### パターン C: トランザクションが必要な複数操作

```go
// adapter に書く
func (a *XxxAdapter) CreateWithRelated(ctx context.Context, ...) error {
    tx, err := a.db.BeginTx(ctx, nil)
    if err != nil { return err }
    defer tx.Rollback()

    qtx := a.q.WithTx(tx)
    qtx.CreateXxx(ctx, ...)
    qtx.CreateRelated(ctx, ...)

    return tx.Commit()
}
```

### パターン D: 認証不要のエンドポイント（マスタデータ等）

```go
// router/skill_router.go のように、グループではなく echo に直接登録
func RegisterXxxRoutes(e *echo.Echo, c *controller.XxxController) {
    e.GET("/xxx", c.ListXxx)  // 認証ミドルウェアなし
}
```

---

## sqlc のよく使う操作

```bash
# SQL クエリを追加・変更したら必ず実行
sqlc generate

# 設定確認
cat sqlc.yaml

# 生成ファイルの確認
ls query/
```

### SQL クエリのアノテーション

```sql
-- name: FunctionName :one    → 1件返す (T, error)
-- name: FunctionName :many   → 複数返す ([]T, error)
-- name: FunctionName :exec   → 件数返さない (error のみ)
```

---

## デバッグのヒント

### ログを確認する

```bash
# Docker でのログ確認
docker compose logs api -f

# Air のリロードログも含まれる
```

### DB に直接接続する

```bash
docker compose exec db psql -U dev_user -d roadmap_dev

# よく使う SQL
\dt              # テーブル一覧
SELECT * FROM users LIMIT 5;
SELECT * FROM user_skills WHERE user_id = 'uuid-here';
```

### JWT のデコード

Supabase の JWT は [jwt.io](https://jwt.io) でデコードできます。`sub` クレームが user_id です。

### よくあるエラー

```
# panic: Echo validator is not registered
→ dicontainer.go に e.Validator = utils.NewValidator() を追加

# sql: no rows in result set (500 になる)
→ errors.Is(err, sql.ErrNoRows) で 404 に分岐

# cannot use string as type uuid.UUID
→ uuid.Parse(str) で変換が必要

# bind error: json: cannot unmarshal
→ リクエストの JSON フィールド名とタグが一致しているか確認
```

---

## マイグレーション

### ローカルでのマイグレーション実行

docker-compose 起動時に `migrate` サービスが自動実行します。

```bash
# 手動実行する場合
docker compose run --rm migrate
```

### 新しいマイグレーションを追加する

```bash
# ファイル名: {番号:6桁}_{説明}.up.sql
# 例:
sql/migrations/000002_add_user_profile.up.sql
```

マイグレーションファイルを追加したら、`sql/schema.sql` にも同様の変更を反映することを忘れずに（sqlc が schema.sql を読むため）。

---

## Git ブランチ運用

コミット履歴を見ると以下のパターンを採用しています：

```
main              本番ブランチ
feature/xxx       機能追加
docs              ドキュメント
```

コミットメッセージは Conventional Commits 形式：

```
feat: 新機能
fix:  バグ修正
docs: ドキュメント
refactor: リファクタリング
```
