# ローカル開発環境設計（Docker Compose）

> Go + PostgreSQL + Redis のローカル環境構築ガイド
> **前提：** Docker Desktop がインストール済みであること

---

## 1. ローカル環境アーキテクチャ

```mermaid
flowchart TB
    subgraph LOCAL["🖥️ ローカル開発環境 (Docker Compose)"]
        direction TB

        subgraph FRONTEND["フロントエンド（ホストで直接起動）"]
            FE["Next.js 14\nlocalhost:3000\nnpm run dev"]
        end

        subgraph CONTAINERS["Dockerコンテナ群"]
            direction LR
            API["api コンテナ\nGo + Air（ホットリロード）\nlocalhost:8080"]
            DB["db コンテナ\nPostgreSQL 15\nlocalhost:5432"]
            REDIS["redis コンテナ\nRedis 7\nlocalhost:6379"]
            MIGRATE["migrate コンテナ\ngolang-migrate\n起動時に自動マイグレーション"]
        end

        FE -->|"API呼び出し"| API
        API -->|"SQL"| DB
        API -->|"キャッシュ"| REDIS
        MIGRATE -->|"マイグレーション実行"| DB
    end

    subgraph EXTERNAL["外部サービス（ローカルでも本物を使用）"]
        CLAUDE["Claude API\nanthropic.com"]
        GITHUB["GitHub\nソースコード管理"]
    end

    API -->|"AI生成リクエスト"| CLAUDE

    style FRONTEND  fill:#dbeafe,stroke:#3b82f6
    style CONTAINERS fill:#d4edda,stroke:#6bcb77
    style EXTERNAL  fill:#fef9c3,stroke:#eab308
```

---

## 2. ディレクトリ構成

```mermaid
flowchart LR
    subgraph ROOT["📁 プロジェクトルート"]
        direction TB
        DC["docker-compose.yml"]
        ENV[".env.local\n（gitignore対象）"]
        ENV_EX[".env.example\n（gitに含める）"]

        subgraph BACKEND["📁 backend/"]
            GO_MOD["go.mod / go.sum"]
            AIR[".air.toml\nホットリロード設定"]
            DOCKERFILE_DEV["Dockerfile.dev\n開発用イメージ"]
            DOCKERFILE_PROD["Dockerfile\n本番用マルチステージ"]

            subgraph CMD["📁 cmd/server/"]
                MAIN["main.go"]
            end

            subgraph INTERNAL["📁 internal/"]
                HANDLER["handler/"]
                SERVICE["service/"]
                REPO["repository/"]
                MW["middleware/"]
                MODEL["model/"]
                WORKER["worker/"]
            end

            subgraph PKG["📁 pkg/"]
                CRYPTO["crypto/\nAES-256"]
                LOGGER["logger/\nzap"]
                VALID["validator/"]
            end

            subgraph DB_DIR["📁 db/"]
                MIGRATIONS["migrations/\n*.sql"]
                QUERIES["queries/\n*.sql（sqlc入力）"]
                SQLC_OUT["sqlc/\n（自動生成Go）"]
                SQLC_YML["sqlc.yaml"]
            end
        end

        subgraph FRONTEND["📁 frontend/"]
            PKG_JSON["package.json"]
            NX_CFG["next.config.js"]
        end
    end

    style ROOT fill:#f8f9fa,stroke:#999
    style BACKEND fill:#dbeafe,stroke:#3b82f6
    style FRONTEND fill:#d4edda,stroke:#6bcb77
```

---

## 3. docker-compose.yml 全体設計

```mermaid
flowchart TD
    subgraph COMPOSE["docker-compose.yml の構成"]
        direction TB

        SVC_API["🟦 api サービス\nimage: backend/Dockerfile.dev\nvolumes: ./backend:/app\nports: 8080:8080\ndepends_on: db, redis\nenvironment: .env.local読み込み\ncommand: air（ホットリロード）"]

        SVC_DB["🟫 db サービス\nimage: postgres:15-alpine\nvolumes: postgres_data:/var/lib/postgresql/data\nports: 5432:5432\nenvironment: POSTGRES_DB/USER/PASSWORD\nhealthcheck: pg_isready"]

        SVC_REDIS["🟥 redis サービス\nimage: redis:7-alpine\nports: 6379:6379\ncommand: redis-server --appendonly yes\nvolumes: redis_data:/data"]

        SVC_MIGRATE["🟨 migrate サービス\nimage: migrate/migrate\ndepends_on: db (healthcheck)\ncommand: -path /migrations -database postgres://... up\nrestart: on-failure"]

        VOLUMES["📦 Named Volumes\npostgres_data\nredis_data"]

        NETWORK["🌐 Network\napp-network (bridge)"]

        SVC_MIGRATE --> SVC_DB
        SVC_API --> SVC_DB
        SVC_API --> SVC_REDIS
        SVC_DB --> VOLUMES
        SVC_REDIS --> VOLUMES
        SVC_API --- NETWORK
        SVC_DB --- NETWORK
        SVC_REDIS --- NETWORK
    end

    style SVC_API    fill:#dbeafe,stroke:#3b82f6
    style SVC_DB     fill:#fef9c3,stroke:#eab308
    style SVC_REDIS  fill:#ffe0e0,stroke:#ff6b6b
    style SVC_MIGRATE fill:#d4edda,stroke:#6bcb77
    style VOLUMES    fill:#ede7f6,stroke:#9b59b6
    style NETWORK    fill:#f0f4f8,stroke:#aaa
```

---

## 4. 環境変数設計

```mermaid
flowchart LR
    subgraph ENV_LOCAL[".env.local（ローカル用・gitignore）"]
        direction TB
        E1["# サーバー\nSERVER_PORT=8080\nAPP_ENV=development"]
        E2["# DB\nDB_HOST=db\nDB_PORT=5432\nDB_NAME=roadmap_dev\nDB_USER=dev_user\nDB_PASSWORD=dev_password\nDB_SSLMODE=disable"]
        E3["# Redis\nREDIS_ADDR=redis:6379\nREDIS_PASSWORD=（空）"]
        E4["# JWT\nJWT_SECRET=local_dev_secret_32chars_xxxx\nJWT_EXPIRES_IN=24h"]
        E5["# AI API（ローカルでは自分のキーを使用）\nCLAUDE_API_KEY=sk-ant-xxxx\nAI_MONTHLY_LIMIT=100000"]
        E6["# 暗号化\nAES_KEY=32bytes_local_dev_key_xxxxxxxx\nAES_IV_PREFIX=local_iv"]
        E7["# CORS\nCORS_ALLOW_ORIGINS=http://localhost:3000"]
    end

    subgraph ENV_EX[".env.example（gitに含める・値はダミー）"]
        direction TB
        EX1["SERVER_PORT=8080\nAPP_ENV=development\nDB_HOST=db\nDB_PORT=5432\nDB_NAME=roadmap_dev\nDB_USER=dev_user\nDB_PASSWORD=CHANGE_ME\nREDIS_ADDR=redis:6379\nJWT_SECRET=CHANGE_ME_32_CHARS\nCLAUDE_API_KEY=CHANGE_ME\nAES_KEY=CHANGE_ME_32_CHARS"]
    end

    subgraph ENV_CLOUD["GCP Secret Manager（本番）"]
        direction TB
        C1["projects/xxx/secrets/jwt-secret\nprojects/xxx/secrets/aes-key\nprojects/xxx/secrets/claude-api-key"]
    end

    ENV_LOCAL -.->|"参考"| ENV_EX
    ENV_LOCAL -.->|"本番は置き換え"| ENV_CLOUD

    style ENV_LOCAL  fill:#ffe0e0,stroke:#ff6b6b
    style ENV_EX     fill:#d4edda,stroke:#6bcb77
    style ENV_CLOUD  fill:#dbeafe,stroke:#3b82f6
```

---

## 5. Air（ホットリロード）設定

```mermaid
flowchart LR
    subgraph AIR[".air.toml 設定内容"]
        direction TB
        A1["root = .\ntmp_dir = tmp"]
        A2["[build]\ncmd = go build -o ./tmp/main ./cmd/server\nbinary = ./tmp/main\nfull_bin = APP_ENV=development ./tmp/main\ninclude_ext = [go, tpl, html]\nexclude_dir = [tmp, vendor, db/sqlc]\ndelay = 1000"]
        A3["[log]\ntime = true"]
        A4["[color]\nmain = magenta\nwatcher = cyan\nbuild = yellow\nrunner = green"]
        A1 --> A2 --> A3 --> A4
    end

    subgraph FLOW["ホットリロードフロー"]
        direction LR
        EDIT["📝 .go ファイル編集・保存"]
        DETECT["👁️ Air がファイル変更を検知"]
        BUILD["⚙️ go build 実行"]
        RESTART["🔄 バイナリ再起動"]
        READY["✅ 変更反映（〜1秒）"]
        EDIT --> DETECT --> BUILD --> RESTART --> READY
    end

    style AIR  fill:#dbeafe,stroke:#3b82f6
    style FLOW fill:#d4edda,stroke:#6bcb77
```

---

## 6. よく使うコマンド

```mermaid
flowchart TB
    subgraph SETUP["🚀 初期セットアップ"]
        direction TB
        S1["# 1. リポジトリクローン後\ncp .env.example .env.local\n# .env.local を編集してAPIキー等を設定"]
        S2["# 2. コンテナ起動（DB・Redis・バックエンド）\ndocker compose up -d"]
        S3["# 3. フロントエンドは別ターミナルで\ncd frontend && npm install && npm run dev"]
        S4["# 4. ブラウザで確認\nhttp://localhost:3000"]
        S1 --> S2 --> S3 --> S4
    end

    subgraph DAILY["📅 日常の開発コマンド"]
        direction TB
        D1["# コンテナ起動・停止\ndocker compose up -d\ndocker compose down"]
        D2["# ログ確認\ndocker compose logs -f api\ndocker compose logs -f db"]
        D3["# DBに直接接続\ndocker compose exec db psql -U dev_user -d roadmap_dev"]
        D4["# マイグレーション追加\n# backend/db/migrations/ に .sql ファイル追加後\ndocker compose restart migrate"]
    end

    subgraph SQLC["🔧 sqlc コード生成"]
        direction TB
        Q1["# backend/db/queries/ に .sql クエリを追加後\ncd backend && sqlc generate\n# backend/db/sqlc/ にGoコードが自動生成される"]
    end

    subgraph RESET["🔄 DB リセット"]
        direction TB
        R1["# 完全リセット（データ消去）\ndocker compose down -v\ndocker compose up -d"]
    end

    style SETUP  fill:#d4edda,stroke:#6bcb77
    style DAILY  fill:#dbeafe,stroke:#3b82f6
    style SQLC   fill:#fef9c3,stroke:#eab308
    style RESET  fill:#ffe0e0,stroke:#ff6b6b
```

---

## 7. Dockerfile.dev（開発用）

```mermaid
flowchart TD
    subgraph DEV_DOCKER["backend/Dockerfile.dev"]
        direction TB
        D1["FROM golang:1.22-alpine\n\n# ビルドツールインストール\nRUN apk add --no-cache git make\n\n# Air（ホットリロード）インストール\nRUN go install github.com/air-verse/air@latest"]

        D2["WORKDIR /app\n\n# 依存関係を先にキャッシュ（レイヤーキャッシュ最適化）\nCOPY go.mod go.sum ./\nRUN go mod download"]

        D3["# ソースはvolumeマウントで反映\n# COPYは不要\n\nEXPOSE 8080\n\nCMD [\"air\", \"-c\", \".air.toml\"]"]

        D1 --> D2 --> D3
    end

    subgraph PROD_DOCKER["backend/Dockerfile（本番・参考）"]
        direction TB
        P1["# Stage 1: ビルド\nFROM golang:1.22-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 GOOS=linux go build\n  -ldflags='-w -s'\n  -o /app/server ./cmd/server"]

        P2["# Stage 2: 実行（最小イメージ）\nFROM gcr.io/distroless/static-debian12\nCOPY --from=builder /app/server /server\nUSER nonroot:nonroot\nEXPOSE 8080\nENTRYPOINT [\"/server\"]"]

        P1 --> P2
    end

    style DEV_DOCKER  fill:#dbeafe,stroke:#3b82f6
    style PROD_DOCKER fill:#d4edda,stroke:#6bcb77
```

---

## 8. ローカル vs Cloud Run の環境差分

```mermaid
flowchart LR
    subgraph LOCAL["🖥️ ローカル環境"]
        direction TB
        L1["DB接続先\ndb:5432\n（Dockerネットワーク内）"]
        L2["Redis接続先\nredis:6379"]
        L3["シークレット管理\n.env.local\n（平文、gitignore）"]
        L4["ホットリロード\nAir使用"]
        L5["TLS\nなし（HTTP）"]
        L6["ログ\nテキスト形式\nコンソール出力"]
    end

    subgraph CLOUD["☁️ Cloud Run（本番）"]
        direction TB
        C1["DB接続先\nCloud SQL Unix Socket\n/cloudsql/proj:region:instance"]
        C2["Redis接続先\nMemorystore\n10.x.x.x:6379（VPC経由）"]
        C3["シークレット管理\nSecret Manager\n環境変数として注入"]
        C4["ホットリロード\n不要（イメージ再デプロイ）"]
        C5["TLS\nCloud Load Balancingで終端"]
        C6["ログ\nJSON形式\nCloud Loggingへ転送"]
    end

    L1 -.->|"APP_ENV で切り替え"| C1
    L2 -.->|"REDIS_ADDR で切り替え"| C2
    L3 -.->|"Secret Managerに移行"| C3
    L6 -.->|"APP_ENV=production でJSON出力"| C6

    style LOCAL fill:#dbeafe,stroke:#3b82f6
    style CLOUD fill:#d4edda,stroke:#6bcb77
```

---

## 9. トラブルシューティング

```mermaid
flowchart TD
    subgraph TROUBLE["よくある問題と解決策"]
        direction TB

        T1["❌ db コンテナが起動しない\n→ docker compose down -v して再起動\n→ ポート5432が他プロセスに使われていないか確認\n   lsof -i :5432"]

        T2["❌ api コンテナが db に接続できない\n→ DB_HOST=db になっているか確認\n   （localhostではなくコンテナ名）\n→ db の healthcheck が通るまで待機"]

        T3["❌ migrate が失敗する\n→ migrationsフォルダのファイル名確認\n   形式: 000001_create_users.up.sql\n→ SQLの構文エラーを確認\n   docker compose logs migrate"]

        T4["❌ Air がホットリロードしない\n→ .air.toml の include_ext を確認\n→ ファイルシステムのパーミッション確認\n→ docker compose restart api"]

        T5["❌ Claude API がローカルで動かない\n→ .env.local の CLAUDE_API_KEY を確認\n→ APIキーが有効か Anthropic Console で確認"]
    end

    style TROUBLE fill:#ffe0e0,stroke:#ff6b6b
```

