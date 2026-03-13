# APIアーキテクチャ設計書

> 参照元：`tech-stack-v3.md` / `db-design.md` / `permission-design-v6.md`

---

## 1. 層状アーキテクチャ（共通）

```mermaid
graph LR
    Client["クライアント\nNext.js / Swagger UI"]
    MW["middleware\nJWT検証 / RBAC / ABAC\nレートリミット"]
    R["router\nルーティング定義"]
    C["controller\nリクエスト/レスポンス変換\nSwagger注釈"]
    S["service\nビジネスロジック"]
    A["adapter\nDB / 外部API通信"]
    EXT["外部サービス\nPostgreSQL / Supabase Auth\nClaude API / Redis"]

    Client -->|"HTTP + Bearer JWT"| MW
    MW --> R
    R --> C
    C --> S
    S --> A
    A --> EXT
```

### 各層の責務

| 層 | ディレクトリ | 責務 |
|----|------------|------|
| middleware | `middleware/` | JWT検証・RBAC/ABACチェック・レートリミット |
| router | `router/` | エンドポイントとcontrollerのマッピング |
| controller | `controller/` | リクエスト/レスポンス変換・Swagger注釈・エラーハンドリング |
| service | `service/` | ビジネスロジック・バリデーション |
| adapter | `adapter/` | DB・外部APIとの通信（sqlcクエリの呼び出し） |
| query | `query/` | sqlcが自動生成した型安全なDBクエリ |

---

## 2. リクエスト処理の詳細フロー

```mermaid
sequenceDiagram
    actor Client as クライアント
    participant MW as middleware
    participant C as controller
    participant S as service
    participant A as adapter
    participant DB as PostgreSQL

    Client->>MW: HTTP Request + Bearer JWT
    MW->>MW: JWT署名検証（SUPABASE_JWT_SECRET）
    MW->>MW: claims.sub → user_id 取得
    MW->>MW: RBAC チェック（global_role）
    MW->>MW: ABAC チェック（チームスコープ等）

    alt 認証・認可NG
        MW-->>Client: 401 / 403
    end

    MW->>C: echo.Context（user_id セット済み）
    C->>C: リクエストボディ バインド＆バリデーション
    C->>S: サービス呼び出し
    S->>A: アダプター呼び出し
    A->>DB: sqlcクエリ実行
    DB-->>A: 結果
    A-->>S: 結果
    S-->>C: ドメインオブジェクト
    C->>C: レスポンス構造体に変換
    C-->>Client: JSON レスポンス
```

---

## 3. 認証フロー（Supabase連携）

```mermaid
sequenceDiagram
    actor User as ユーザー
    participant FE as Next.js
    participant SUPA as Supabase Auth
    participant GO as Go API
    participant DB as PostgreSQL

    User->>FE: メール・パスワード入力
    FE->>SUPA: supabase.auth.signInWithPassword()
    SUPA-->>FE: access_token（JWT）+ refresh_token

    FE->>GO: POST /auth/login\nAuthorization: Bearer JWT
    GO->>GO: Supabase JWT検証
    GO->>DB: EnsureUser（user_profilesにupsert）
    GO->>DB: AssignGlobalRole（LOGIN_USER付与）
    GO-->>FE: 200 OK + ユーザー情報

    Note over FE,GO: 以降のAPIリクエスト
    FE->>GO: GET /users/me\nAuthorization: Bearer JWT
    GO->>GO: middleware でJWT検証・user_id 取得
    GO->>DB: SELECT FROM user_profiles
    GO-->>FE: ユーザープロフィール
```

---

## 4. 開発環境

### 構成図

```mermaid
flowchart TB
    subgraph Local["ローカルマシン"]
        subgraph Docker["docker-compose"]
            API["api コンテナ\nGo + Air（ホットリロード）\nDockerfile.dev\n:8080"]
            DB["db コンテナ\nPostgreSQL 15\n:5432"]
            Redis["redis コンテナ\nRedis 7\n:6379"]
            Migrate["migrate コンテナ\ngolang-migrate\nsql/migrations/ を適用"]
        end
        SRC["ソースコード\n./ → /app にvolume mount"]
    end

    subgraph Supabase["Supabase（クラウド）"]
        SUPA_AUTH["Supabase Auth\nJWT発行・メール確認"]
    end

    Browser["ブラウザ\nSwagger UI\nhttp://localhost:8080/swagger/"]

    Browser --> API
    API --> DB
    API --> Redis
    API --> SUPA_AUTH
    Migrate --> DB
    SRC -.->|"volume mount"| API
```

### 起動方法

```bash
# コンテナ起動（migrate も自動実行）
docker compose up

# マイグレーションのみ再実行
docker compose run --rm migrate

# sqlcコード再生成
make sqlc

# Swagger docs再生成
swag init -g main.go
```

### 開発環境の特徴

| 項目 | 内容 |
|------|------|
| ホットリロード | Air（`.air.toml`）でソース変更を即反映 |
| DB | ローカルPostgreSQL（`roadmap_dev`） |
| マイグレーション | `sql/migrations/` を golang-migrate で管理 |
| Supabase Auth | 本番Supabaseに接続（JWT検証はクラウドへ） |
| Swagger UI | `http://localhost:8080/swagger/` で有効 |
| 環境変数 | `.env.local` から読み込み |

### ディレクトリ構成（開発関連）

```
roadmap_api/
├── Dockerfile.dev          # 開発用（Air入り）
├── docker-compose.yml      # ローカル環境一式
├── .env.local              # 開発用環境変数（git管理外）
├── sql/
│   ├── migrations/         # golang-migrate マイグレーションファイル
│   ├── queries/            # sqlc用クエリ定義（*.sql）
│   └── schema.sql          # スキーマ全体
├── query/                  # sqlcが自動生成（編集禁止）
├── supabase/
│   └── migrations/         # Supabase CLIマイグレーション（参考用）
└── docs/                   # swag initで自動生成（編集禁止）
```

---

## 5. 本番環境

### 構成図

```mermaid
flowchart TB
    subgraph GCP["Google Cloud Platform"]
        LB["Cloud Load Balancing\nHTTPS終端・SSL証明書"]
        CR["Cloud Run\nGo API コンテナ\nDockerfile マルチステージビルド\ncpu:1 / mem:512Mi\nmin:0 / max:10 オートスケール"]
        AR["Artifact Registry\nDockerイメージ管理"]
        SM["Secret Manager\nSUPABASE_JWT_SECRET\nSUPABASE_SERVICE_KEY\nCLAUDE_API_KEY\nAES_KEY"]
        MEM["Memorystore Redis\nレートリミット"]
        LOG["Cloud Logging\nMonitoring"]
    end

    subgraph Supabase["Supabase（SaaS）"]
        SUPA_AUTH["Auth\nJWT発行・OAuth管理"]
        SUPA_DB["PostgreSQL\nアプリDB"]
        SUPA_ST["Storage\n画像ファイル"]
    end

    subgraph Frontend["フロントエンド"]
        Vercel["Vercel\nNext.js"]
    end

    subgraph CICD["CI/CD"]
        GHA["GitHub Actions\ndocker build → AR push → Cloud Run deploy"]
    end

    Vercel -->|"HTTPS + Bearer JWT"| LB
    Vercel <-->|"認証フロー"| SUPA_AUTH
    LB --> CR
    CR --> SM
    CR --> MEM
    CR --> SUPA_DB
    CR --> SUPA_AUTH
    CR --> LOG
    AR --> CR
    GHA --> AR
    GHA --> Vercel
```

### 起動方法（CI/CD）

```bash
# GitHub Actions が mainブランチへのマージで自動実行
# 1. docker build（マルチステージ）
# 2. Artifact Registry へ push
# 3. Cloud Run へ deploy
```

### 本番環境の特徴

| 項目 | 内容 |
|------|------|
| コンテナ | `Dockerfile`（distroless, nonrootユーザー）|
| DB | Supabase PostgreSQL（マネージド） |
| マイグレーション | Supabase CLI / ダッシュボードで管理 |
| シークレット | Google Cloud Secret Manager から注入 |
| Swagger UI | `APP_ENV=production` のとき無効化推奨 |
| スケール | Cloud Run オートスケール（min:0） |
| Redis | Memorystore（またはUpstash無料枠） |

---

## 6. 開発環境 vs 本番環境 比較

| 項目 | 開発環境 | 本番環境 |
|------|----------|----------|
| 起動方法 | `docker compose up` | GitHub Actions（CI/CD） |
| Dockerfile | `Dockerfile.dev`（Air入り） | `Dockerfile`（distroless） |
| DB | ローカルPostgreSQL（Docker） | Supabase PostgreSQL |
| マイグレーション | `sql/migrations/`（golang-migrate） | Supabase CLIまたはダッシュボード |
| Redis | Dockerコンテナ | Memorystore / Upstash |
| 環境変数 | `.env.local` | Cloud Secret Manager |
| Swagger UI | 有効（`/swagger/`） | 無効化推奨 |
| ホットリロード | Air（有効） | なし |
| ログ | 標準出力 | Cloud Logging |
| イメージサイズ | 大（Alpine + Air + devtools） | 小（distroless, ~10MB程度） |

---

## 7. テスト戦略

```mermaid
flowchart LR
    subgraph TestTargets["テスト対象層"]
        S["service\nビジネスロジック"]
        A["adapter\nDBクエリ・外部API"]
    end

    subgraph Mock["モック（tests/mock/）"]
        MS["MockAdapterInterface\nadapterをモック化"]
        ME["MockExternalService\n外部サービスをモック化"]
    end

    subgraph Tools["ツール"]
        MG["mockgen\nmake mockgen で生成"]
        TF["testify/assert\nアサーション"]
    end

    S -->|"interfaceに依存"| A
    MS -.->|"serviceテスト時に注入"| S
    ME -.->|"adapterテスト時に注入"| A
    MG -->|"interface → mock生成"| Mock
```

### テスト実行

```bash
# モック再生成
make mockgen

# テスト実行
go test ./...

# カバレッジ確認
go test ./... -cover
```

### テスト方針

- **controller はテスト対象外**（リクエスト/レスポンス変換のみのため）
- **service** はアダプターをモック化してビジネスロジックを検証
- **adapter** はDBを実際に使った統合テストを推奨（モックとの乖離防止）
