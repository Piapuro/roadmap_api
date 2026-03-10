# 権限設計書 v6

> 達成目標：最小権限 / 誤付与防止 / 監査可能性 / 運用可能性
> **v6変更点：RBAC（基本）+ ABAC（例外・細粒度制御）のハイブリッド設計を追加。**

---

## 1. RBAC × ABAC ハイブリッド設計の全体像

```mermaid
flowchart TB
    subgraph HYBRID["🏗️ ハイブリッド権限モデル"]
        direction LR

        subgraph RBAC["🟦 RBAC（基本・土台）\nRole-Based Access Control"]
            RB1["ロール単位で権限を定義\n覚えやすく管理しやすい"]
            RB2["5ロールで全体をカバー\nSYSTEM_ADMIN / TEAM_OWNER\nTEAM_MEMBER / LOGIN_USER / GUEST"]
            RB3["チームスコープで独立管理\n複数チーム所属に対応"]
        end

        subgraph ABAC["🟨 ABAC（例外・細粒度）\nAttribute-Based Access Control"]
            AB1["属性の組み合わせで制御\nロールだけでは表現できない条件"]
            AB2["このプロダクトでの属性例\nチームレベル / スキル区分\nロードマップ状態 / タスクオーナー"]
            AB3["ルール管理コストが高いため\n例外ケースのみに限定適用"]
        end

        RBAC -->|"RBACで許可されたうえで\nABACで追加条件チェック"| ABAC
    end

    JUDGE["⚖️ 最終判定\nRBAC ✅ かつ ABAC ✅ → アクセス許可\nRBAC ❌ → 即ブロック（ABACは見ない）\nRBAC ✅ かつ ABAC ❌ → 条件不一致でブロック"]

    HYBRID --> JUDGE

    style RBAC   fill:#dbeafe,stroke:#3b82f6
    style ABAC   fill:#fef9c3,stroke:#eab308
    style HYBRID fill:#f8f9fa,stroke:#999
    style JUDGE  fill:#d4edda,stroke:#6bcb77
```

---

## 2. このプロダクトで使う属性一覧（ABAC）

```mermaid
mindmap
  root((🏷️ ABAC 属性一覧))
    👤 ユーザー属性
      skill_level
        beginner 初心者
        intermediate 中級
        advanced 上級
      assigned_tasks_count
        タスク担当数
        負荷分散チェックに使用
      is_team_owner
        そのチームのOWNERか
      joined_at
        チーム参加日時
    📦 リソース属性
      team_level
        初心者チーム
        混合チーム
        上級チーム
      roadmap_status
        draft 未確定
        confirmed 確定済み
        archived アーカイブ
      task_owner_id
        タスクの担当者ID
      requirement_status
        draft 作成中
        locked 確定ロック済み
    🌍 環境属性
      request_time
        アクセス日時
      is_team_member
        そのチームに所属しているか
      team_id
        操作対象のチームID
```

---

## 3. ABAC ルール定義（例外ケース）

```mermaid
flowchart TD
    subgraph RULE1["📌 ルール1：ロードマップ確定後の要件定義ロック"]
        R1C["条件\nroadmap_status = confirmed\nかつ\nユーザーロール = TEAM_MEMBER"]
        R1A["結果\n要件定義の編集・削除 → ❌ ブロック\n※ RBACでは TEAM_MEMBERは元々編集不可\n　 OWNERも確定後は編集不可にしたい場合に使用"]
        R1C --> R1A
    end

    subgraph RULE2["📌 ルール2：自分のタスクのみステータス変更可"]
        R2C["条件\ntask_owner_id = 操作ユーザーのuser_id\nかつ\nユーザーロール = TEAM_MEMBER"]
        R2A["結果\n自分のタスクのみ更新 → ✅ 許可\n他人のタスク更新 → ❌ ブロック\n※ RBACだけでは「自分のタスクのみ」を表現できない"]
        R2C --> R2A
    end

    subgraph RULE3["📌 ルール3：チームレベルによるAI再生成制限"]
        R3C["条件\nteam_level = beginner\nかつ\nユーザーロール = TEAM_MEMBER"]
        R3A["結果\nAI再生成 → ❌ ブロック\n（初心者チームはOWNERのみ再生成可）\n※ 初心者の誤操作・コスト爆発防止"]
        R3C --> R3A
    end

    subgraph RULE4["📌 ルール4：タスク過負荷ブロック"]
        R4C["条件\nassigned_tasks_count >= 5\nかつ\nユーザーロール = TEAM_OWNER（担当割り振り時）"]
        R4A["結果\n追加アサイン → ⚠️ 警告表示\n（強制ブロックではなく警告）\n※ できる人への負荷集中を防ぐ"]
        R4C --> R4A
    end

    subgraph RULE5["📌 ルール5：チームスコープ境界の強制"]
        R5C["条件\nrequest_team_id ≠ user所属チームID\nかつ\nユーザーロール = TEAM_OWNER or TEAM_MEMBER"]
        R5A["結果\n他チームのリソース操作 → ❌ 403ブロック\n存在自体を見せない\n※ 複数チーム所属時の越境アクセス防止"]
        R5C --> R5A
    end

    style RULE1 fill:#dbeafe,stroke:#3b82f6
    style RULE2 fill:#fef9c3,stroke:#eab308
    style RULE3 fill:#ffe0e0,stroke:#ff6b6b
    style RULE4 fill:#d4edda,stroke:#6bcb77
    style RULE5 fill:#ede7f6,stroke:#9b59b6
```

---

## 4. 権限判定フロー（リクエストごとの処理）

```mermaid
flowchart TD
    REQ(["📨 APIリクエスト"])

    REQ --> AUTH["🔐 Step1. 認証チェック\nJWT / Session検証"]
    AUTH -->|"❌ 未認証"| E401["401 Unauthorized\nログイン画面へ"]
    AUTH -->|"✅ 認証OK"| RBAC_CHECK

    RBAC_CHECK["🟦 Step2. RBACチェック\n必要ロール以上か？\nチームスコープ一致か？"]
    RBAC_CHECK -->|"❌ ロール不足"| E403["403 Forbidden\n操作不可メッセージ"]
    RBAC_CHECK -->|"✅ RBAC通過"| ABAC_NEEDED

    ABAC_NEEDED{"🤔 Step3. ABAC対象か？\n例外ルールに該当する\n操作か判定"}
    ABAC_NEEDED -->|"対象外\n通常操作"| ALLOW["✅ アクセス許可\nリソース返却"]
    ABAC_NEEDED -->|"対象\n例外ルールあり"| ABAC_CHECK

    ABAC_CHECK["🟨 Step4. ABACチェック\n属性条件を評価\nルール1〜5を順次チェック"]
    ABAC_CHECK -->|"❌ 条件不一致"| E403B["403 Forbidden\n条件不一致メッセージ"]
    ABAC_CHECK -->|"⚠️ 警告条件"| WARN["⚠️ 警告付きで許可\n（過負荷警告など）"]
    ABAC_CHECK -->|"✅ 全条件クリア"| ALLOW

    LOG["📋 Step5. 監査ログ記録\nWHO / WHAT / WHEN\nRBAC結果 / ABACルールID / RESULT"]
    ALLOW --> LOG
    WARN  --> LOG
    E403  --> LOG
    E403B --> LOG

    style REQ   fill:#f8f9fa,stroke:#999
    style AUTH  fill:#dbeafe,stroke:#3b82f6
    style RBAC_CHECK fill:#dbeafe,stroke:#3b82f6
    style ABAC_NEEDED fill:#fef9c3,stroke:#eab308
    style ABAC_CHECK fill:#fef9c3,stroke:#eab308
    style ALLOW fill:#d4edda,stroke:#6bcb77
    style WARN  fill:#fff3cd,stroke:#ffc107
    style LOG   fill:#ede7f6,stroke:#9b59b6
    style E401  fill:#ffe0e0,stroke:#ff6b6b
    style E403  fill:#ffe0e0,stroke:#ff6b6b
    style E403B fill:#ffe0e0,stroke:#ff6b6b
```

---

## 5. RBAC vs ABAC 使い分けマップ

```mermaid
quadrantChart
    title RBAC vs ABAC 使い分け（制御の細かさ × 管理コスト）
    x-axis 管理コスト低い --> 管理コスト高い
    y-axis 制御が粗い --> 制御が細かい
    quadrant-1 ABACで対応
    quadrant-2 理想だが過剰
    quadrant-3 RBACで対応
    quadrant-4 運用限界・使わない
    ログイン・ログアウト: [0.1, 0.1]
    チーム作成: [0.15, 0.2]
    メンバー一覧閲覧: [0.2, 0.25]
    要件定義入力: [0.25, 0.3]
    AI生成実行: [0.3, 0.35]
    ロードマップ確定: [0.35, 0.4]
    タスク作成・削除: [0.2, 0.3]
    自タスク更新のみ許可: [0.65, 0.75]
    確定後の編集ロック: [0.6, 0.7]
    チームスコープ境界: [0.55, 0.8]
    過負荷警告アサイン: [0.7, 0.65]
    初心者チームAI制限: [0.6, 0.6]
```

---

## 6. DB設計（ABAC対応追加）

```mermaid
flowchart LR
    subgraph RBAC_TABLES["🟦 RBACテーブル（既存）"]
        direction TB
        U["users\nid / email / name\nskill_level\npassword_hash"]
        GR["global_roles\nid / name / level"]
        UGR["user_global_roles\nuser_id / global_role_id"]
        T["teams\nid / name / level\nis_public"]
        TR["team_roles\nid / name / level"]
        UTR["user_team_roles\nuser_id / team_id / team_role_id"]
        U --- UGR --- GR
        U --- UTR --- TR
        T --- UTR
    end

    subgraph ABAC_TABLES["🟨 ABACテーブル（追加）"]
        direction TB
        AR["abac_rules\nid / name / description\ntarget_action\ncondition_json\neffect(allow/deny/warn)\npriority"]
        AL["abac_rule_logs\nid / rule_id / user_id\nteam_id / action\nattributes_snapshot\nresult / created_at"]
        AR --- AL
    end

    NOTE["💡 condition_jsonの例（ルール2）\n{\n  user.role: TEAM_MEMBER,\n  resource.task_owner_id:\n    != user.id\n  → effect: deny\n}"]

    AR -.- NOTE

    style RBAC_TABLES fill:#dbeafe,stroke:#3b82f6
    style ABAC_TABLES fill:#fef9c3,stroke:#eab308
    style NOTE        fill:#d4edda,stroke:#6bcb77
```

---

## 7. 設計サマリー（RBAC + ABAC）

```mermaid
flowchart LR
    subgraph G1["🟦 RBACで解決すること"]
        G1A["ロール単位の大枠の権限管理\n覚えやすく・管理しやすい"]
        G1B["5ロール構成\nSYSTEM_ADMIN / TEAM_OWNER\nTEAM_MEMBER / LOGIN_USER / GUEST"]
        G1C["チームスコープで\n複数チーム所属に対応"]
    end

    subgraph G2["🟨 ABACで解決すること"]
        G2A["自分のタスクのみ更新可\n→ task_owner_id = user.id"]
        G2B["ロードマップ確定後の編集ロック\n→ roadmap_status = confirmed"]
        G2C["初心者チームのAI操作制限\n→ team_level = beginner"]
        G2D["タスク過負荷の警告\n→ assigned_tasks_count >= 5"]
        G2E["チームスコープ越境ブロック\n→ team_id 不一致"]
    end

    subgraph G3["⚖️ 判定の優先順位"]
        G3A["① 認証チェック（401）"]
        G3B["② RBACチェック（403）"]
        G3C["③ ABACチェック（403 or ⚠️）"]
        G3D["RBACで弾かれたら\nABACは評価しない\nパフォーマンス最適化"]
        G3A --> G3B --> G3C
        G3B -.- G3D
    end

    subgraph G4["📋 運用コスト管理"]
        G4A["基本はRBAC\nほぼ全機能はロールで解決"]
        G4B["ABACは5ルールのみ\n例外に絞って管理コストを抑制"]
        G4C["ABACルールはDBで管理\n管理画面から変更可能"]
    end

    style G1 fill:#dbeafe,stroke:#3b82f6
    style G2 fill:#fef9c3,stroke:#eab308
    style G3 fill:#d4edda,stroke:#6bcb77
    style G4 fill:#ffe0e0,stroke:#ff6b6b
```
