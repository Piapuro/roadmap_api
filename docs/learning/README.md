# 📚 コードベース学習ガイド

このドキュメントは [code-learner skill] によって自動生成されました。

**対象プロジェクト**: roadmap_api（Go / Echo / PostgreSQL / Supabase）
**生成日時**: 2026-03-09

---

## 読む順番

| # | ファイル | 内容 | 読む目安 |
|---|---------|------|---------|
| 1 | [00_overview.md](00_overview.md) | プロジェクト全体図・技術スタック・API一覧 | まずここから |
| 2 | [01_architecture.md](01_architecture.md) | 4層アーキテクチャ・DI・ミドルウェア設計 | 構造を理解する |
| 3 | [02_data_flow.md](02_data_flow.md) | リクエストから DB まで追うシーケンス図 | 動きを理解する |
| 4 | [03_key_concepts.md](03_key_concepts.md) | sqlc・トランザクション・Null値・バリデーション | パターンを学ぶ |
| 5 | [04_getting_started.md](04_getting_started.md) | セットアップ・新機能追加の手順 | 手を動かす |

---

## 5分でわかる概要

```
クライアント
    ↓ Authorization: Bearer <JWT>
Middleware（Supabase JWT 検証）
    ↓ ctx.Set("user_id", ...)
Controller（Bind + Validate + HTTP ステータス決定）
    ↓
Service（ビジネスロジック・型変換）
    ↓
Adapter（トランザクション管理・sqlc 呼び出し）
    ↓
query/*.go（sqlc 自動生成・手書き禁止）
    ↓
PostgreSQL
```

**全依存は `dicontainer/dicontainer.go` 一箇所で組み立てています。**

---

## よく使うコマンド

```bash
# 開発環境起動
docker compose up --build

# SQL クエリを変更したら
sqlc generate

# ビルド確認
go build ./...
```
