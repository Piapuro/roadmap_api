# AI API コスト・セキュリティ設計

---

## 1. APIキー管理の3モデル比較

```mermaid
flowchart TB
    subgraph M1["🅰️ モデル1：サービス側で一元管理（SaaS型）"]
        direction LR
        M1U["ユーザー"] -->|"リクエスト"| M1S["サービスサーバー\n（APIキーを保持）"]
        M1S -->|"APIキーで呼び出し"| M1AI["Claude API\nOpenAI API"]
        M1N["✅ ユーザーはAPIキー不要\n✅ セキュリティ管理が一箇所\n✅ 体験がシンプル\n❌ 運営がAPI費用を全額負担\n❌ スパム・乱用リスクが運営に集中"]
    end

    subgraph M2["🅱️ モデル2：チーム・ユーザーが自分のAPIキーを登録（BYOK型）"]
        direction LR
        M2U["ユーザー\n（自分のAPIキーを登録）"] -->|"リクエスト"| M2S["サービスサーバー\n（キーを暗号化保存）"]
        M2S -->|"登録キーで呼び出し"| M2AI["Claude API\nOpenAI API"]
        M2N["✅ 運営のAPI費用ゼロ\n✅ コスト責任がユーザー側\n⚠️ キーの安全な保管が必要\n⚠️ キー漏洩リスクをどう管理するか\n❌ ユーザーがAPIキー取得の手間あり\n❌ 初心者には設定ハードル高い"]
    end

    subgraph M3["🅲️ モデル3：ハイブリッド（無料枠＋BYOK上乗せ）"]
        direction LR
        M3U["ユーザー"] -->|"まず無料枠を使用"| M3S["サービスサーバー"]
        M3U -.->|"枠を超えたら\n自分のAPIキー使用"| M3S
        M3S -->|"状況に応じて\nキーを切り替え"| M3AI["Claude API\nOpenAI API"]
        M3N["✅ 初心者はすぐ使える\n✅ ヘビーユーザーはBYOKで拡張\n✅ 運営コストに上限を設けられる\n⚠️ 実装複雑度が上がる\n⚠️ 無料枠の範囲設計が重要"]
    end

    style M1 fill:#dbeafe,stroke:#3b82f6
    style M2 fill:#fef9c3,stroke:#eab308
    style M3 fill:#d4edda,stroke:#6bcb77
```

---

## 2. BYOK（Bring Your Own Key）採用時のセキュリティリスクと対策

```mermaid
flowchart TD
    subgraph RISKS["🚨 BYOKの主なリスク"]
        direction TB
        R1["⚠️ リスク1\nAPIキーの平文保存\nDBが漏洩したとき全キーが露出"]
        R2["⚠️ リスク2\nサーバーログへの混入\nリクエストログにキーが残る"]
        R3["⚠️ リスク3\nチーム内での横断利用\n他メンバーが直接キーを見られる"]
        R4["⚠️ リスク4\nキーの無制限利用\nサーバー経由で無制限に呼べてしまう"]
        R5["⚠️ リスク5\nキー漏洩後の影響範囲\n1チームのキー漏洩が他チームに波及しないか"]
    end

    subgraph MEASURES["🛡️ 対策"]
        direction TB
        S1["✅ 対策1：保存時暗号化\nAES-256でDBに暗号化保存\n表示時はマスキング（sk-****xxxx）\n復号はAPI呼び出し時のみ・メモリ上で即破棄"]
        S2["✅ 対策2：ログのサニタイズ\nリクエスト・レスポンスログから\nAPIキーを自動マスク\nログ基盤への混入を防ぐ"]
        S3["✅ 対策3：キーの間接利用\nユーザー・チームメンバーは\nキーの実値を取得できない\nサーバー経由でのみ使用"]
        S4["✅ 対策4：呼び出し制限\nチームごとに月次トークン上限を設定\n上限超過で自動停止・通知"]
        S5["✅ 対策5：チーム完全分離\nキーはチームスコープで管理\n他チームのキーへのアクセスは\n物理的に不可能な設計"]
    end

    R1 --> S1
    R2 --> S2
    R3 --> S3
    R4 --> S4
    R5 --> S5

    style RISKS    fill:#ffe0e0,stroke:#ff6b6b
    style MEASURES fill:#d4edda,stroke:#6bcb77
```

---

## 3. 推奨アーキテクチャ：ハイブリッド型の詳細設計

```mermaid
flowchart TD
    USER(["👤 ユーザー（チームメンバー）"])

    USER -->|"AI生成リクエスト"| GATE

    subgraph GATE["🔐 API Gateway / サーバー層"]
        direction TB
        G1["① 認証・権限チェック\nJWT検証 + RBAC + ABAC"]
        G2{"② キー選択ロジック\nどのAPIキーを使うか"}
        G3["③ トークン消費チェック\n今月の使用量 vs 上限"]
        G1 --> G2 --> G3
    end

    subgraph KEY_SELECT["🔑 キー選択の優先順位"]
        direction TB
        K1["優先1️⃣：チーム登録BYOKキー\nそのチームのオーナーが登録したキー"]
        K2["優先2️⃣：サービス共有キー（無料枠）\n月Xトークンまで運営が負担"]
        K3["優先3️⃣：上限超過 → エラー\n「BYOKキーを登録してください」へ誘導"]
        K1 --> K2 --> K3
    end

    G2 --> KEY_SELECT

    subgraph SECURE_STORE["🗄️ セキュアキーストア"]
        direction LR
        KS1["service_api_key\n運営のAPIキー\n環境変数で管理\nコードに埋め込まない"]
        KS2["team_api_keys テーブル\nteam_id / encrypted_key\n/ monthly_limit / used_tokens\n→ AES-256暗号化保存"]
    end

    KEY_SELECT --> SECURE_STORE

    SECURE_STORE -->|"復号してリクエスト\n使用後メモリから破棄"| CLAUDE["🤖 Claude API"]

    CLAUDE -->|"レスポンス"| LOG
    LOG["📋 使用量ログ記録\nteam_id / tokens_used\ntimestamp / action_type\n※ APIキー実値は記録しない"]
    LOG --> USER

    style GATE         fill:#dbeafe,stroke:#3b82f6
    style KEY_SELECT   fill:#fef9c3,stroke:#eab308
    style SECURE_STORE fill:#ffe0e0,stroke:#ff6b6b
    style CLAUDE       fill:#d4edda,stroke:#6bcb77
    style LOG          fill:#ede7f6,stroke:#9b59b6
```

---

## 4. トークン使用量管理・コスト制御

```mermaid
flowchart LR
    subgraph LIMITS["📊 使用量制限の設計"]
        direction TB
        L1["🔴 サービス共有枠\n月 X万トークンまで無料\n（運営が設定・変更可能）\n超過 → BYOKへ誘導"]
        L2["🟠 チームBYOK枠\nチームOWNERが月次上限を設定\n超過 → 自動停止 + OWNER通知"]
        L3["🟢 アクション別重み付け\nAI分析実行：重い（多トークン）\n再生成：中程度\nタスク単体確認：軽い"]
        L1 --> L2 --> L3
    end

    subgraph MONITOR["📈 コスト監視"]
        direction TB
        MON1["チームダッシュボードに表示\n今月の使用トークン数\n残りトークン数・上限まで何%"]
        MON2["SYSTEM_ADMINダッシュボード\n全チームの使用量一覧\n異常検知アラート\n（短時間に大量消費など）"]
        MON3["OWNER向けアラート通知\n80%到達 → 警告通知\n100%到達 → 停止通知\nBYOK登録の案内"]
    end

    subgraph ABUSE["🚫 乱用防止"]
        direction TB
        AB1["レートリミット\n同一チームから1分間X回まで"]
        AB2["アクション単位の上限\nAI再生成は1日Y回まで\n（初心者チームはOWNERのみ）"]
        AB3["異常検知\n深夜の大量リクエスト\n同一内容の連続リクエストを検知"]
    end

    style LIMITS  fill:#dbeafe,stroke:#3b82f6
    style MONITOR fill:#d4edda,stroke:#6bcb77
    style ABUSE   fill:#ffe0e0,stroke:#ff6b6b
```

---

## 5. BYOKキー登録フロー（TEAM_OWNERのみ）

```mermaid
flowchart TD
    OWNER(["🟠 TEAM_OWNER"])

    OWNER -->|"チーム設定画面\nAPIキー登録"| INPUT["APIキー入力フォーム\nsk-ant-xxxxxxxxxxxx"]

    INPUT --> VALIDATE["バリデーション\n① フォーマットチェック\n② Anthropic/OpenAIへの\n   テストリクエスト（1トークン）\n   → 有効なキーか確認"]

    VALIDATE -->|"❌ 無効"| ERR["エラー表示\n無効なAPIキーです"]
    VALIDATE -->|"✅ 有効"| ENCRYPT["AES-256で暗号化\nencrypted_key として保存\n実値は即メモリから削除"]

    ENCRYPT --> SAVE["team_api_keys に保存\n表示は sk-****...xxxx のみ\n実値は二度と画面に出ない"]

    SAVE --> LIMIT["月次トークン上限を設定\nデフォルト：500,000 tokens\nOWNERが変更可能"]

    LIMIT --> DONE["✅ 登録完了\nこのチームのAI機能が\nBYOKキーで動作開始"]

    subgraph RULE["⚠️ BYOKのルール"]
        R1["登録・変更・削除は\nTEAM_OWNERのみ"]
        R2["TEAM_MEMBERはキーの\n存在は知れるが実値は見えない"]
        R3["SYSTEM_ADMINも\n暗号化された値しか見えない"]
        R4["キー削除後は\nサービス共有枠に自動フォールバック"]
    end

    style RULE fill:#fff8e1,stroke:#ffc107
```

---

## 6. モデル比較サマリー

```mermaid
quadrantChart
    title 3モデルの比較（コスト負担 × 導入ハードル）
    x-axis ユーザー導入ハードル低い --> ユーザー導入ハードル高い
    y-axis 運営コスト負担小 --> 運営コスト負担大
    quadrant-1 運営負担大・使いやすい（持続困難）
    quadrant-2 運営負担大・使いにくい（最悪）
    quadrant-3 運営負担小・使いにくい（ユーザー離脱）
    quadrant-4 運営負担小・使いやすい（理想）
    モデル1 SaaS型: [0.2, 0.85]
    モデル2 完全BYOK: [0.8, 0.1]
    モデル3 ハイブリッド推奨: [0.35, 0.35]
```

---

## 7. 結論・推奨設計

```mermaid
flowchart LR
    subgraph REC["✅ 推奨：ハイブリッド型（モデル3）"]
        direction TB
        P1["📌 フェーズ1 MVP\nサービス共有枠のみ\n月Xトークンまで運営負担\nシンプルに始める"]
        P2["📌 フェーズ2 本格運用\nBYOK登録機能を追加\nチームOWNERが自分のキーを登録\n共有枠を超えたらBYOKへ自動切替"]
        P3["📌 フェーズ3 拡張\nトークン使用量ダッシュボード\n異常検知アラート\nレートリミット強化"]
        P1 --> P2 --> P3
    end

    subgraph SECURITY["🛡️ 必須セキュリティ対応"]
        direction TB
        S1["APIキーはAES-256暗号化保存"]
        S2["ログへのキー混入を防ぐサニタイズ"]
        S3["チームスコープで完全分離"]
        S4["TEAM_OWNERのみキー操作可"]
        S5["レートリミット・月次上限の設定"]
    end

    style REC      fill:#d4edda,stroke:#6bcb77
    style SECURITY fill:#dbeafe,stroke:#3b82f6
```
