# 00 プロジェクト全体図

## 何を作っているか

**roadmap_api** は、チーム開発の学習を支援する **ロードマップ生成サービス** のバックエンドAPIです。

ユーザーが自分の技術スキルや学習目標を登録すると、AIがチームに最適な開発ロードマップ（Epic・Story・Task の構成）を自動生成します。

```
ユーザー登録 → スキル入力 → チーム結成 → 要件定義 → AI がロードマップ生成 → タスク管理
```

---

## 技術スタック

| 用途 | 技術 | バージョン |
|------|------|-----------|
| 言語 | Go | 1.24 |
| Web フレームワーク | Echo | v4.15 |
| データベース | PostgreSQL | 15 |
| 認証 | Supabase JWT | - |
| クエリ生成 | sqlc | v1.30 |
| AI | Google Gemini API | - |
| キャッシュ | Redis | 7 |
| バリデーション | go-playground/validator | v10 |
| コンテナ | Docker / Docker Compose | - |
| ホットリロード | Air | v1.61 |

---

## ディレクトリ構造

```
roadmap_api/
├── main.go                    # エントリポイント（2行だけ）
│
├── dicontainer/               # 依存性注入コンテナ
│   └── dicontainer.go         # 全依存関係の初期化・配線
│
├── driver/                    # 外部サービスへの接続設定
│   ├── psql.go                # PostgreSQL 接続
│   ├── supabase.go            # Supabase 設定読み込み
│   └── gemini.go              # Google Gemini API クライアント
│
├── middleware/                # HTTP ミドルウェア
│   ├── supabase_auth.go       # JWT 認証（実装済み）
│   ├── rbac.go                # ロールベース認可（TODO）
│   └── abac.go                # 属性ベース認可（TODO）
│
├── router/                    # ルーティング定義
│   ├── auth_router.go         # /auth/*
│   ├── user_router.go         # /users/*
│   ├── team_router.go         # /teams/*
│   ├── requirement_router.go  # /requirements/*
│   ├── roadmap_router.go      # /roadmaps/*
│   ├── webhook_router.go      # /webhooks/*
│   └── skill_router.go        # /skills
│
├── controller/                # HTTP ハンドラー（リクエスト受付・レスポンス返却）
│   ├── auth_controller.go
│   ├── user_controller.go
│   ├── team_controller.go
│   ├── requirement_controller.go
│   ├── roadmap_controller.go
│   ├── webhook_controller.go
│   └── skill_controller.go
│
├── service/                   # ビジネスロジック層
│   ├── auth_service.go
│   ├── user_service.go
│   ├── team_service.go
│   ├── requirement_service.go
│   ├── ai_service.go          # Gemini 呼び出しロジック
│   └── roadmap_service.go
│
├── adapter/                   # DB 操作層（sqlc クエリのラッパー）
│   ├── user_adapter.go
│   ├── team_adapter.go
│   ├── requirement_adapter.go
│   ├── roadmap_adapter.go
│   ├── ai_adapter.go
│   └── webhook_adapter.go
│
├── query/                     # sqlc が自動生成した DAO コード（手書き禁止）
│   ├── db.go                  # Queries 構造体・WithTx
│   ├── models.go              # 全テーブルの Go 型定義
│   ├── users.sql.go           # users/user_skills クエリ
│   └── teams.sql.go           # teams クエリ
│
├── requests/                  # リクエストボディの型定義
│   ├── user_request.go
│   ├── skill_request.go
│   ├── team_request.go
│   ├── requirement_request.go
│   └── roadmap_request.go
│
├── response/                  # レスポンスボディの型定義
│   ├── user_response.go
│   ├── skill_response.go
│   ├── team_response.go
│   ├── requirement_response.go
│   └── roadmap_response.go
│
├── utils/                     # 共通ユーティリティ
│   ├── validator.go           # Echo 向けバリデータ実装
│   └── error_util.go          # エラーレスポンス生成関数
│
├── sql/                       # SQL ソースファイル（sqlc の入力）
│   ├── schema.sql             # テーブル定義（CREATE TABLE）
│   ├── queries/               # sqlc が読む SQL クエリ
│   │   ├── users.sql
│   │   └── teams.sql
│   └── migrations/            # DB マイグレーション
│       └── 000001_init.up.sql
│
├── Dockerfile                 # 本番ビルド（distroless、nonroot）
├── Dockerfile.dev             # 開発用（Air ホットリロード）
├── docker-compose.yml         # ローカル開発環境一式
└── sqlc.yaml                  # sqlc 設定
```

---

## APIエンドポイント一覧

```
# 認証不要
GET  /health                            ヘルスチェック
GET  /skills                            スキルタグマスタ
POST /auth/signup                       ユーザー登録
POST /auth/login                        ログイン
POST /auth/logout                       ログアウト
POST /webhooks/supabase/user-created    Supabase Webhook

# 認証必須（Authorization: Bearer <JWT>）
GET  /users/me                          自分のプロフィール取得
PUT  /users/me                          プロフィール更新
GET  /users/me/skills                   自分のスキル取得
PUT  /users/me/skills                   スキル一括登録・更新

POST   /teams                           チーム作成
GET    /teams                           チーム一覧
GET    /teams/:id                       チーム詳細
PUT    /teams/:id                       チーム更新
DELETE /teams/:id                       チーム削除

POST /requirements                      要件定義作成
GET  /requirements/:id                  要件定義取得
PUT  /requirements/:id                  要件定義更新
POST /requirements/:id/submit           要件定義確定

POST   /roadmaps                        ロードマップ生成（AI）
GET    /roadmaps                        ロードマップ一覧
GET    /roadmaps/:id                    ロードマップ詳細
PUT    /roadmaps/:id                    ロードマップ更新
DELETE /roadmaps/:id                    ロードマップ削除
```

---

## システム全体のデータフロー（俯瞰）

```
┌─────────────────────────────────────────────────────────┐
│                      外部クライアント                       │
│              (フロントエンド / モバイルアプリ)               │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTPS + JWT
                            ▼
┌─────────────────────────────────────────────────────────┐
│                     Echo Web Server                      │
│  ┌──────────────┐    ┌──────────────┐   ┌────────────┐  │
│  │  Middleware   │ → │  Controller  │ → │  Service   │  │
│  │  JWT認証      │    │  Bind/Validate│   │ ビジネス   │  │
│  │  RBAC(予定)   │    │  Response組立 │   │ ロジック   │  │
│  └──────────────┘    └──────────────┘   └─────┬──────┘  │
│                                               │         │
│                                         ┌─────▼──────┐  │
│                                         │  Adapter   │  │
│                                         │  DB操作    │  │
│                                         └─────┬──────┘  │
└───────────────────────────────────────────────┼─────────┘
                                                │
              ┌─────────────────────────────────┼──────────┐
              │                                 │          │
              ▼                                 ▼          ▼
    ┌──────────────────┐             ┌───────────────┐  ┌──────┐
    │    PostgreSQL     │             │ Gemini API    │  │Redis │
    │  (Supabase)       │             │ (AI生成)      │  │(予定)│
    └──────────────────┘             └───────────────┘  └──────┘
```

---

## 実装状況（2026年3月時点）

| 機能 | 状態 |
|------|------|
| スキル登録・取得 | ✅ 完全実装 |
| Supabase Webhook でロール自動付与 | ✅ 完全実装 |
| JWT 認証ミドルウェア | ✅ 完全実装 |
| バリデーション基盤 | ✅ 完全実装 |
| Docker 開発環境 | ✅ 完全実装 |
| 認証（サインアップ/ログイン） | 🚧 骨組みのみ |
| チーム管理 | 🚧 骨組みのみ |
| 要件定義 | 🚧 骨組みのみ |
| AI ロードマップ生成 | 🚧 骨組みのみ |
| RBAC / ABAC 認可 | 🚧 骨組みのみ |
