# 技術構成設計書 v3
---

## 1. 技術スタック全体像

```mermaid
flowchart TB
    subgraph CLIENT["🖥️ クライアント層"]
        direction LR
        FE["Next.js 14\nReact 18 + TypeScript\nTailwind CSS + shadcn/ui"]
        STATE["Zustand\n状態管理"]
        SSE["Server-Sent Events\nAIストリーミング表示"]
        DND["dnd-kit\nカンバンDnD"]
        SUPA_CLIENT["@supabase/ssr\n認証状態管理\nJWT取得"]
        FE --- STATE
        FE --- SSE
        FE --- DND
        FE --- SUPA_CLIENT
    end

    subgraph SERVER["⚙️ サーバー層（Go）"]
        direction LR
        API["Echo\nGo 1.22"]
        MW["ミドルウェア\nSupabase JWT検証\nRBAC / ABAC\nレートリミット"]
        WORKER["Goroutine Worker\nAI非同期処理\nチャネル通信"]
        API --> MW --> WORKER
    end

    subgraph SUPABASE["🔷 Supabase"]
        direction LR
        AUTH["Auth\nメール / Google OAuth\nJWT発行・管理"]
        SUPA_DB["PostgreSQL\nアプリDBも同居\nRow Level Security"]
        STORAGE["Storage\nアイコン画像など"]
    end

    subgraph AI["🤖 AI層"]
        CLAUDE["Claude API\nGo HTTPクライアント\nSSEストリーミング"]
        KEYSTORE["Secret Manager\nBYOKキー管理\nAES-256暗号化"]
        CLAUDE --- KEYSTORE
    end

    subgraph INFRA["☁️ Google Cloud"]
        direction LR
        CR["Cloud Run\nコンテナ実行\nオートスケール"]
        AR["Artifact Registry\nDockerイメージ管理"]
        LB["Cloud Load Balancing\nHTTPS終端"]
        GHA["GitHub Actions\nCI/CD"]
        CR --- AR --- LB --- GHA
    end

    CLIENT  -->|"HTTPS / REST + JWT"| LB
    CLIENT  <-->|"認証フロー"| SUPABASE
    LB      --> CR
    CR      --> SERVER
    SERVER  -->|"supabase-go / REST"| SUPABASE
    SERVER  -->|"HTTP"| AI

    style CLIENT   fill:#dbeafe,stroke:#3b82f6
    style SERVER   fill:#d4edda,stroke:#16a34a
    style SUPABASE fill:#ede7f6,stroke:#9b59b6
    style AI       fill:#ffe0e0,stroke:#ff6b6b
    style INFRA    fill:#e8f5e9,stroke:#4caf50
```


---

## 3. 認証フロー（Supabase）

```mermaid
sequenceDiagram
    actor User as 👤 ユーザー
    participant FE as 🖥️ Next.js
    participant SUPA as 🔷 Supabase Auth
    participant GO as ⚙️ Go + Echo
    participant DB as 🗃️ Supabase DB

    Note over User, DB: メール/パスワード登録・ログイン
    User->>FE: メールアドレス・パスワード入力
    FE->>SUPA: supabase.auth.signUp() / signIn()
    SUPA-->>FE: JWT（access_token）+ refresh_token
    FE->>FE: JWTをlocalStorageまたはCookieに保存

    Note over User, DB: APIリクエスト
    User->>FE: 操作（チーム作成など）
    FE->>GO: POST /api/teams\nAuthorization: Bearer <JWT>
    GO->>GO: SupabaseのJWT_SECRETで検証\n（Supabase公開鍵 or 共有シークレット）
    GO->>GO: claims.Sub → user_id 取得
    GO->>GO: RBAC / ABACチェック
    GO->>DB: チームデータ保存（Supabase PostgreSQL）
    DB-->>GO: 保存完了
    GO-->>FE: 201 Created

    Note over User, DB: Google OAuth
    User->>FE: Googleでログインボタン
    FE->>SUPA: supabase.auth.signInWithOAuth({provider: 'google'})
    SUPA->>SUPA: Googleリダイレクト処理
    SUPA-->>FE: JWT（自動発行） ← OAuth実装ほぼゼロ 🎉
```

---

## 4. Go側 Supabase JWT検証（シンプルになる）

```mermaid
flowchart TB
    subgraph GO_AUTH["⚙️ GoのJWT検証（v3）"]
        direction TB

        OLD["❌ v2 自前実装\n・bcryptでパスワード検証\n・JWTを自前で発行・署名\n・refreshトークン管理\n・OAuthフロー実装\n→ 約 30h"]

        NEW["✅ v3 Supabase検証のみ\n・SupabaseのJWT_SECRETで署名検証\n・claims.Sub を user_id として使用\n・それ以外はSupabaseが全部やってくれる\n→ 約 2〜3h"]

        CODE["// middleware/jwt.go\nfunc AuthMiddleware() echo.MiddlewareFunc {\n  return func(next echo.HandlerFunc) echo.HandlerFunc {\n    return func(c echo.Context) error {\n      tokenStr := extractBearerToken(c)\n      \n      // Supabaseの公開鍵/シークレットで検証するだけ\n      token, err := jwt.Parse(tokenStr,\n        func(t *jwt.Token) (interface{}, error) {\n          return []byte(os.Getenv(\"SUPABASE_JWT_SECRET\")), nil\n        })\n      \n      if err != nil {\n        return c.JSON(401, map[string]string{\"error\": \"unauthorized\"})\n      }\n      \n      claims := token.Claims.(jwt.MapClaims)\n      c.Set(\"user_id\", claims[\"sub\"]) // Supabaseのuser.id\n      return next(c)\n    }\n  }\n}"]
    end

    OLD -.->|"Supabaseに置き換え"| NEW
    NEW --> CODE

    style OLD  fill:#ffe0e0,stroke:#ff6b6b
    style NEW  fill:#d4edda,stroke:#16a34a
    style CODE fill:#f8f9fa,stroke:#aaa
```

---

## 5. DBスキーマ設計（Supabase PostgreSQL）

```mermaid
erDiagram
    auth_users["auth.users（Supabase管理）"] {
        uuid id PK
        string email
        jsonb raw_user_meta_data
        timestamp created_at
    }
    user_profiles["user_profiles（アプリ管理）"] {
        uuid id PK
        uuid auth_user_id FK "auth.users.idと紐づけ"
        string name
        string skill_level
        string[] tech_tags
        timestamp created_at
    }
    global_roles {
        uuid id PK
        string name
        int level
    }
    user_global_roles {
        uuid user_id FK
        uuid global_role_id FK
    }
    teams {
        uuid id PK
        string name
        string description
        string level
        boolean is_public
        timestamp created_at
    }
    team_roles {
        uuid id PK
        string name
        int level
    }
    user_team_roles {
        uuid user_id FK
        uuid team_id FK
        uuid team_role_id FK
    }
    requirements {
        uuid id PK
        uuid team_id FK
        string product_type
        jsonb feature_checklist
        string difficulty
        text free_text
        string status
        timestamp created_at
    }
    roadmaps {
        uuid id PK
        uuid team_id FK
        string status
        timestamp confirmed_at
        timestamp created_at
    }
    phases {
        uuid id PK
        uuid roadmap_id FK
        int phase_number
        string title
    }
    tasks {
        uuid id PK
        uuid phase_id FK
        string title
        text description
        string status
        uuid task_owner_id FK
        int estimated_hours
        timestamp due_date
        timestamp created_at
    }
    task_comments {
        uuid id PK
        uuid task_id FK
        uuid user_id FK
        text content
        timestamp created_at
    }
    team_api_keys {
        uuid id PK
        uuid team_id FK
        bytea encrypted_key
        int monthly_limit
        int used_tokens
        timestamp created_at
    }
    api_usage_logs {
        uuid id PK
        uuid team_id FK
        uuid user_id FK
        string action_type
        int tokens_used
        timestamp created_at
    }
    audit_logs {
        uuid id PK
        uuid user_id FK
        uuid team_id FK
        string action
        string resource_type
        string result
        jsonb attributes_snapshot
        timestamp created_at
    }

    auth_users ||--|| user_profiles : "auth_user_id"
    user_profiles ||--o{ user_global_roles : ""
    global_roles ||--o{ user_global_roles : ""
    user_profiles ||--o{ user_team_roles : ""
    teams ||--o{ user_team_roles : ""
    team_roles ||--o{ user_team_roles : ""
    teams ||--|| requirements : ""
    teams ||--|| roadmaps : ""
    roadmaps ||--o{ phases : ""
    phases ||--o{ tasks : ""
    tasks ||--o{ task_comments : ""
    tasks }o--|| user_profiles : ""
    teams ||--o| team_api_keys : ""
    teams ||--o{ api_usage_logs : ""
    user_profiles ||--o{ audit_logs : ""
```

> **ポイント：** `auth.users` はSupabaseが管理。アプリ側は `user_profiles` テーブルで独自データを管理し、`auth_user_id` で紐づける。

---

## 6. フロントエンド認証実装（Next.js + Supabase）

```mermaid
flowchart TB
    subgraph FE_AUTH["🖥️ Next.js 認証実装"]
        direction TB

        SETUP["// lib/supabase.ts\nimport { createBrowserClient } from '@supabase/ssr'\n\nexport const supabase = createBrowserClient(\n  process.env.NEXT_PUBLIC_SUPABASE_URL!,\n  process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!\n)"]

        LOGIN["// メールログイン（数行で完成）\nconst { data, error } = await supabase.auth.signInWithPassword({\n  email,\n  password,\n})"]

        OAUTH["// Google OAuth（1行で完成）\nawait supabase.auth.signInWithOAuth({\n  provider: 'google',\n  options: { redirectTo: '/dashboard' }\n})"]

        API_CALL["// GoのAPIを叩くとき\nconst { data: { session } } = await supabase.auth.getSession()\n\nfetch('/api/teams', {\n  headers: {\n    Authorization: `Bearer ${session?.access_token}`\n  }\n})"]

        SETUP --> LOGIN --> OAUTH --> API_CALL
    end

    style FE_AUTH fill:#dbeafe,stroke:#3b82f6
```

---

## 7. Go パッケージ構成（v3）

```mermaid
mindmap
  root((📦 Go パッケージ v3))
    Webフレームワーク
      labstack/echo
        高速HTTPルーター
        ミドルウェア対応
      swaggo/swag
        コメントからSwagger自動生成
        swag init で openapi.json 出力
      swaggo/echo-swagger
        /swagger/* エンドポイント提供
        Swagger UI をEchoに組み込み
    認証・セキュリティ
      golang-jwt/jwt
        Supabase JWT検証のみ
        発行はSupabaseが担当
      crypto/aes
        AES-256 BYOKキー暗号化
        標準ライブラリ
    Supabase連携
      supabase-community/supabase-go
        Supabase REST API クライアント
        DBアクセスにも使用可能
    DB・クエリ
      sqlc（推奨）
        SQL→Goコード自動生成
        Supabase DBに直接接続
      lib/pq
        PostgreSQLドライバ
      golang-migrate/migrate
        マイグレーション管理
    Redis
      go-redis/redis
        レートリミット
        キャッシュ（Memorystore）
    バリデーション
      go-playground/validator
    ロギング
      uber-go/zap
        構造化ログ・JSON出力
    AI連携
      net/http 標準ライブラリ
        Claude APIクライアント
        SSEストリーミング
    設定管理
      spf13/viper
        環境変数管理
    テスト
      testify/assert
      httptest
```

---

## 8. Swagger（OpenAPI）設計

```mermaid
flowchart TB
    subgraph SWAGGER["📄 Swagger / OpenAPI 構成"]
        direction TB

        subgraph SETUP["セットアップ"]
            S1["swaggo/swag\nコメントアノテーションから\nopenapi.json を自動生成\n$ swag init -g cmd/server/main.go"]
            S2["swaggo/echo-swagger\nEchoルーターに\nSwagger UIを組み込み\nGET /swagger/*"]
            S1 --> S2
        end

        subgraph ANNOTATION["コメントアノテーション例"]
            A1["// main.go\n// @title           Roadmap Dashboard API\n// @version         1.0\n// @description     サークル開発ロードマップ自動生成API\n// @host            localhost:8080\n// @BasePath        /api\n// @securityDefinitions.apikey BearerAuth\n// @in header\n// @name Authorization"]

            A2["// handler/team.go\n// @Summary      チーム作成\n// @Description  新しいチームを作成しOWNERロールを付与\n// @Tags         teams\n// @Accept       json\n// @Produce      json\n// @Param        body body CreateTeamRequest true \"チーム情報\"\n// @Success      201  {object} TeamResponse\n// @Failure      400  {object} ErrorResponse\n// @Failure      401  {object} ErrorResponse\n// @Security     BearerAuth\n// @Router       /teams [post]"]
        end

        subgraph ROUTES["Swagger UIアクセス"]
            R1["開発環境\nhttp://localhost:8080/swagger/index.html"]
            R2["本番環境\n※ APP_ENV=production のとき\nSwagger UI を無効化推奨\n（機密情報漏洩防止）"]
        end

        subgraph FLOW["自動生成フロー"]
            direction LR
            F1["① Goハンドラーに\nコメント記述"]
            F2["② $ swag init\ndocs/ フォルダに\nopenapi.json 生成"]
            F3["③ echo-swagger が\n/swagger/* で\nSwagger UI を配信"]
            F4["④ フロントチームが\nAPIドキュメントを参照\n手動同期不要"]
            F1 --> F2 --> F3 --> F4
        end
    end

    style SWAGGER    fill:#fef9c3,stroke:#eab308
    style SETUP      fill:#fffbeb,stroke:#f59e0b
    style ANNOTATION fill:#f8f9fa,stroke:#aaa
    style ROUTES     fill:#dbeafe,stroke:#3b82f6
    style FLOW       fill:#d4edda,stroke:#16a34a
```

---

## 10. AI生成フロー（変更なし・Goroutine + SSE）

```mermaid
sequenceDiagram
    actor User as 👤 ユーザー
    participant FE as 🖥️ Next.js
    participant GO as ⚙️ Go + Echo
    participant CH as 📬 Job Channel
    participant WK as 🔧 Goroutine Worker
    participant SM as 🔑 Secret Manager
    participant CL as 🤖 Claude API
    participant DB as 🗃️ Supabase DB

    User->>FE: AI分析実行
    FE->>GO: POST /api/ai/generate\nAuthorization: Bearer <Supabase JWT>

    GO->>GO: Supabase JWT検証
    GO->>GO: RBAC / ABACチェック
    GO->>DB: 要件・メンバースキル取得
    GO->>CH: jobをchannelに送信
    GO-->>FE: 202 Accepted { job_id }

    FE->>GO: GET /api/ai/stream/{job_id}
    GO-->>FE: SSE接続確立

    CH->>WK: goroutineでjobを受信
    WK->>SM: APIキー取得（BYOK or 共有枠）
    SM-->>WK: 復号済みキー

    WK->>CL: HTTPリクエスト（stream=true）
    loop ストリーミング中
        CL-->>WK: チャンク（delta）
        WK-->>FE: SSE data: {chunk}
        FE->>FE: UIにリアルタイム表示
    end

    WK->>DB: 生成結果を保存
    WK->>DB: トークン使用量を記録
    WK->>WK: キーをメモリから即破棄
    WK-->>FE: SSE event: complete
```

---

## 9. インフラ構成（v3）

```mermaid
flowchart TB
    subgraph GCP["☁️ Google Cloud Platform"]
        direction TB
        CR_API["Cloud Run\nGo APIサービス\ncpu:1 mem:512Mi\nmin:0 max:10"]
        AR["Artifact Registry\nDockerイメージ"]
        SM["Secret Manager\nSUPABASE_JWT_SECRET\nSUPABASE_SERVICE_KEY\nCLAUDE_API_KEY\nAES_KEY"]
        MEM["Memorystore Redis\nレートリミット"]
        LOG["Cloud Logging / Monitoring"]
        CR_API --> SM
        CR_API --> MEM
        AR --> CR_API
        CR_API --> LOG
    end

    subgraph SUPABASE["🔷 Supabase（SaaS）"]
        AUTH2["Supabase Auth\nJWT発行・OAuth管理"]
        SUPA_DB2["Supabase PostgreSQL\nアプリDB"]
        SUPA_ST["Supabase Storage\n画像ファイル"]
    end

    subgraph FRONTEND2["🖥️ フロントエンド"]
        VERCEL2["Vercel\nNext.js ホスティング"]
    end

    subgraph CICD2["🔄 CI/CD"]
        GHA2["GitHub Actions\ndocker build → AR → Cloud Run"]
    end

    VERCEL2 -->|"Supabase JWT付きリクエスト"| GCP
    VERCEL2 <-->|"認証フロー"| SUPABASE
    GCP <-->|"DB操作"| SUPABASE
    CICD2 --> GCP
    CICD2 --> VERCEL2

    style GCP       fill:#e8f5e9,stroke:#4caf50
    style SUPABASE  fill:#ede7f6,stroke:#9b59b6
    style FRONTEND2 fill:#dbeafe,stroke:#3b82f6
    style CICD2     fill:#fef9c3,stroke:#eab308

    note1["💡 v2との違い：\nCloud SQLが不要 → Supabase DBに統合\nJWT発行不要 → Supabaseが担当\nOAuth実装不要 → Supabase Authが担当"]
```

---

## 11. 環境変数設計（v3）

```mermaid
flowchart LR
    subgraph ENV["環境変数一覧（v3）"]
        direction TB

        E1["# Supabase\nSUPABASE_URL=https://xxx.supabase.co\nSUPABASE_ANON_KEY=eyJ...\nSUPABASE_SERVICE_KEY=eyJ... ← 管理用（Secret Manager）\nSUPABASE_JWT_SECRET=your-jwt-secret ← JWT検証用"]

        E2["# AI\nCLAUDE_API_KEY=sk-ant-xxx ← Secret Manager\nAI_MONTHLY_LIMIT=100000"]

        E3["# 暗号化（BYOKキー保護）\nAES_KEY=32bytes_key ← Secret Manager"]

        E4["# Redis（レートリミット）\nREDIS_ADDR=10.x.x.x:6379"]

        E5["# サーバー\nSERVER_PORT=8080\nAPP_ENV=production\nCORS_ALLOW_ORIGINS=https://your-app.vercel.app"]

        E6["# フロントエンド（Next.js）\nNEXT_PUBLIC_SUPABASE_URL=https://xxx.supabase.co\nNEXT_PUBLIC_SUPABASE_ANON_KEY=eyJ..."]
    end

    subgraph REMOVED["❌ v2から削除された環境変数"]
        R1["DB_HOST / DB_PORT / DB_USER / DB_PASSWORD\nJWT_SECRET（Supabase_JWT_SECRETに統合）\nBCRYPT_COST（bcrypt不要）"]
    end

    style ENV     fill:#d4edda,stroke:#16a34a
    style REMOVED fill:#ffe0e0,stroke:#ff6b6b
```

---

## 12. Supabase 無料枠・コスト

```mermaid
flowchart TB
    subgraph SUPABASE_FREE["🔷 Supabase 無料枠（Free Plan）"]
        direction TB
        F1["✅ DB：500MB\nサークル規模で十分"]
        F2["✅ Auth：50,000 MAU\nサークルなら余裕"]
        F3["✅ Storage：1GB\nアイコン画像程度なら十分"]
        F4["✅ Realtime：あり"]
        F5["✅ プロジェクト数：2つ\n（開発・本番で使い切る）"]
        F6["⚠️ 制限：7日間非アクティブでDB停止\n本番は Pro（$25/月）推奨"]
    end

    subgraph COST_TOTAL["💰 v3 月額コスト概算（MVP期）"]
        direction TB
        C1["Supabase Free Plan：$0\n（本番移行時 Pro $25/月）"]
        C2["Cloud Run：ほぼ$0\n（無料枠 200万リクエスト/月）"]
        C3["Memorystore Redis：約$15/月\n※ Upstash Redis無料枠で代替可能\n→ $0に"]
        C4["Secret Manager：$0\n（月1万アクセスまで無料）"]
        C5["Vercel Free：$0"]
        C_TOT["MVP期合計：\n約 $0〜$15/月 🎉\n（Redis代替で$0も可能）"]
        C1 & C2 & C3 & C4 & C5 --> C_TOT
    end

    style SUPABASE_FREE fill:#ede7f6,stroke:#9b59b6
    style COST_TOTAL    fill:#d4edda,stroke:#16a34a
```

---

## 13. 技術選定サマリー（v3）

```mermaid
mindmap
  root((🛠️ 技術スタック v3))
    🖥️ フロントエンド
      Next.js 14 + TypeScript
      Tailwind CSS + shadcn/ui
      Zustand 状態管理
      dnd-kit カンバンDnD
      @supabase/ssr 認証クライアント
    ⚙️ バックエンド Go
      Go 1.22 + Echo
      swaggo/swag + echo-swagger Swagger UI
      golang-jwt JWT検証のみ
      crypto/aes AES-256
      sqlc 型安全クエリ生成
      golang-migrate マイグレーション
      go-redis レートリミット
      uber-go/zap 構造化ログ
      Goroutine + Channel AI非同期
    🔷 Supabase
      Auth メール・Google OAuth・JWT
      PostgreSQL アプリDB
      Storage アイコン画像
    🤖 AI
      Claude API Go HTTPクライアント
      SSE http.Flusher
      Secret Manager BYOK管理
      AES-256暗号化
    ☁️ Google Cloud
      Cloud Run コンテナ実行
      Artifact Registry イメージ管理
      Secret Manager シークレット管理
      Memorystore Redis
        or Upstash Redis 無料枠
      Cloud Logging Monitoring
    🔄 CI/CD
      GitHub Actions
      Docker マルチステージビルド
      Vercel フロントデプロイ
```

