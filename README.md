# Connpass Discord 通知システム - 要件定義書

## 📚 目次

1. [プロジェクト概要](#1-プロジェクト概要)
2. [システム全体構成](#2-システム全体構成)
3. [主な機能一覧](#3-主な機能一覧)
4. [通知条件の種類](#4-通知条件の種類)
5. [技術スタック詳細](#5-技術スタック詳細)
6. [データベース設計](#6-データベース設計)
7. [REST API 設計](#7-rest-api-設計)
8. [Discord Bot権限設定](#8-discord-bot権限設定)
9. [Discord OAuth2認証フロー](#9-discord-oauth2認証フロー)
10. [スケジューラ動作詳細](#10-スケジューラ動作詳細)
11. [エラーログ管理](#11-エラーログ管理)
12. [ディレクトリ構成](#12-ディレクトリ構成)
13. [デプロイ構成](#13-デプロイ構成)
14. [開発環境セットアップ](#14-開発環境セットアップ)
15. [セキュリティ・運用](#15-セキュリティ運用)
16. [コスト試算](#16-コスト試算)
17. [今後の拡張予定](#17-今後の拡張予定)

---

## 1. プロジェクト概要

### 🎯 目的

イベント管理サービス「**connpass**」に掲載されたイベント情報を定期的に監視し、指定した条件に合致したイベントを**Discord**へ自動通知するシステムを構築する。

### 🌟 特徴

- **Web画面で簡単設定**: Next.jsで構築されたダッシュボードから、キーワード・地域・通知タイミングを設定
- **自動監視**: 30分ごとにconnpass APIを呼び出し、新規イベントや条件変更を検出
- **重複通知防止**: データベースで送信履歴を管理し、同じイベントを何度も通知しない
- **Discord OAuth2認証**: Discordアカウントでログインし、所属サーバーのルールのみ管理可能
- **チャンネル自動作成**: Web画面からDiscordチャンネルを作成可能

---

## 2. システム全体構成

### 📦 コンポーネント一覧

| コンポーネント | 技術 | デプロイ先 | 役割 |
|----------------|------|------------|------|
| **フロントエンド** | Next.js 15 (App Router) | Vercel | 通知ルールの設定UI、ダッシュボード |
| **API サーバー** | Go 1.25.2 + Echo | Railway (常駐) | REST API提供、認証処理 |
| **Discord Bot** | Go 1.25.2 + discordgo | Railway (常駐) | Discord通知送信、チャンネル管理 |
| **スケジューラ** | Go 1.25.2 (同一コード) | Railway Cron Job | 30分ごとにconnpass API呼び出し |
| **データベース** | PostgreSQL 16 | Railway | データ永続化 |

**補足**: バックエンドの3つ(API・Bot・Scheduler)は**同一のGoプロジェクト**で、起動コマンドで動作を切り替えます。

---

## 3. 主な機能一覧

### ✅ コア機能

| 機能名 | 概要 | 実装優先度 |
|--------|------|-----------|
| **Discord OAuth2認証** | Discordアカウントでログイン可能 | 🔴 必須 |
| **通知ルール作成** | キーワード・地域・通知タイミングを設定 | 🔴 必須 |
| **イベント自動監視** | 30分ごとにconnpass APIを呼び出し | 🔴 必須 |
| **Discord通知送信** | 条件に合致したイベントをDiscordへ通知 | 🔴 必須 |
| **重複通知防止** | 送信済みイベントは再送しない | 🔴 必須 |
| **サーバー別権限管理** | 所属Discordサーバーのルールのみ操作可能 | 🔴 必須 |
| **チャンネル一覧取得** | Botが参加しているチャンネルをプルダウン表示 | 🟡 推奨 |
| **チャンネル作成** | Web画面からDiscordチャンネルを新規作成 | 🟡 推奨 |
| **手動通知テスト** | 指定チャンネルにテストメッセージを送信 | 🟢 オプション |
| **ステータス監視画面** | スケジューラの実行状態・エラー履歴を表示 | 🟡 推奨 |

---

## 4. 通知条件の種類

### 🔔 通知トリガー

| トリガー名 | 説明 | 判定方法 |
|------------|------|----------|
| **新規公開 (open)** | connpassにイベントが新規追加されたとき | `events_cache`にevent_idが存在しない |
| **申込開始 (start)** | イベントの受付開始日時を迎えたとき | `started_at`が現在時刻の前後30分以内 |
| **残席わずか (almost_full)** | 参加率が指定閾値を超えたとき | `(accepted / limit) * 100 >= threshold` |
| **締切前 (before_deadline)** | イベント終了の1時間前になったとき | `ended_at - 1時間`が現在時刻の前後30分以内 |

**注意**:
- 各ルールで複数の通知タイミングを組み合わせ可能（例: 新規公開 + 締切前）
- スケジューラは30分ごとに実行されるため、判定タイミングに±30分の誤差が発生する可能性あり

---

## 5. 技術スタック詳細

### 🛠️ バックエンド 

- Go 1.25.2 linux/amd64


#### 主要ライブラリ

- github.com/labstack/echo/v4 // Webフレームワーク
- github.com/bwmarrin/discordgo // Discord API
- github.com/lib/pq // PostgreSQL driver
- github.com/golang-jwt/jwt/v5 // JWT認証
- github.com/joho/godotenv // 環境変数管理
- golang.org/x/oauth2 // OAuth2クライアント


#### アーキテクチャパターン

**Clean Architecture 風の3層構造**:
```
Handler層 (HTTPリクエスト処理)
    ↓
Service層 (ビジネスロジック)
    ↓
Repository層 (DB操作)
```

---

### 🎨 フロントエンド

- Next.js

---

### 🗄️ データベース
- PostgreSQL

---

## 6. データベース設計

### 📊 テーブル一覧

| テーブル名 | 役割 | レコード数目安 |
|------------|------|---------------|
| **users** | Discord OAuth2でログインしたユーザー情報 | 〜1,000 |
| **guilds** | Botが参加しているDiscordサーバー情報 | 〜100 |
| **rules** | 通知ルール設定 | 〜500 |
| **rule_keywords** | ルールごとの検索キーワード | 〜1,000 |
| **rule_notify_types** | ルールごとの通知タイミング | 〜1,000 |
| **events_cache** | connpassから取得したイベント情報のキャッシュ | 〜10,000 |
| **notifications** | 送信済み通知履歴（重複防止用） | 〜50,000 |
| **important_logs** | 重要なエラーログ・イベントログ | 〜10,000 |
| **scheduler_status** | スケジューラの実行状態管理 | 1 |

---

## 7. REST API 設計

### 🌐 エンドポイント一覧

#### 認証系

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| POST | `/api/auth/callback` | Discord OAuth2コールバック | 不要 |
| POST | `/api/auth/logout` | ログアウト | 必要 |
| GET | `/api/auth/me` | 現在のユーザー情報取得 | 必要 |

#### ルール管理

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| GET | `/api/rules` | ルール一覧取得 | 必要 |
| GET | `/api/rules/:id` | ルール詳細取得 | 必要 |
| POST | `/api/rules` | ルール作成 | 必要 |
| PUT | `/api/rules/:id` | ルール更新 | 必要 |
| DELETE | `/api/rules/:id` | ルール削除 | 必要 |

#### Discord管理

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| GET | `/api/guilds` | ユーザーが所属するギルド一覧 | 必要 |
| GET | `/api/guilds/:id/channels` | ギルドのチャンネル一覧 | 必要 |
| POST | `/api/guilds/:id/channels` | チャンネル作成 | 必要 |
| POST | `/api/test-notification` | テスト通知送信 | 必要 |

#### ステータス監視

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| GET | `/api/status` | システムステータス取得 | 必要 |
| GET | `/api/logs` | エラーログ取得 | 必要 |

---

## 8. Discord Bot権限設定

### Bot招待URL

```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=2148005952&scope=bot
```

### 権限の設定

Web上で各サーバーごとに設定できるようにする。


### 🤖 Bot設定手順

1. **Discord Developer Portal**にアクセス: https://discord.com/developers/applications
2. 「New Application」をクリックしてアプリケーション作成
3. 左メニュー「Bot」→「Add Bot」
4. **Privileged Gateway Intents**を設定:
   - ✅ `SERVER MEMBERS INTENT` (ギルドメンバー情報取得)
   - ✅ `MESSAGE CONTENT INTENT` (将来的な拡張用)
5. 「OAuth2」→「URL Generator」で招待URLを生成:
   - Scopes: `bot`
   - Bot Permissions: 上記の権限を選択
6. 生成されたURLでDiscordサーバーに招待

---

## 9. Discord OAuth2認証フロー

### 🔑 認証の流れ

1. ユーザーがフロントエンドの「Discordでログイン」ボタンをクリック
2. Discord認証画面にリダイレクト（スコープ: `identify guilds`）
   - `identify`: ユーザーID、ユーザー名、アバターの取得
   - `guilds`: ユーザーが所属するサーバー一覧と権限情報の取得
3. ユーザーが認証を許可すると、`code`パラメータ付きでコールバックURLへリダイレクト
4. フロントエンドがGo APIの `/api/auth/callback` に `code` を送信
5. Go APIがDiscord APIで `code` を `access_token` に交換
6. Go APIがDiscord APIからユーザー情報とギルド一覧を取得
   - 各ギルドの権限ビット（`permissions`）も取得される
   - 権限ビットから管理者権限やチャンネル管理権限を判定
7. Go APIがユーザー情報をDBに保存（既存ユーザーは更新）
8. Go APIがJWTトークンを生成して返却
9. フロントエンドがJWTをHttpOnly Cookieに保存し、ダッシュボードへリダイレクト

### 🛡️ サーバー別権限管理

**権限チェックの流れ**:

1. ユーザーがルールを作成・編集しようとする
2. APIがJWTからユーザーIDを取得
3. ユーザーが操作対象のDiscordサーバーに所属しているか確認
4. そのサーバー内での権限を確認:
   - **管理者権限** (0x8ビット) → すべての操作が可能
   - **チャンネル管理権限** (0x10ビット) → ルール作成・チャンネル作成が可能
5. 権限がない場合は403エラーを返却

---

## 10. スケジューラ動作詳細

### ⏰ 実行タイミング

**Cron設定**: `0,30 * * * *`（毎時0分・30分に実行）

**実行時間の目安**:
- 10ルール × 3キーワード = 30 API呼び出し
- 1回の呼び出し = 約2秒（レート制限対策の待機時間含む）
- **合計実行時間**: 約1分

### 🔒 排他制御（アドバイザリロック）

**PostgreSQLのアドバイザリロック**を使用して、複数のスケジューラが同時実行されないようにします。

### 📡 connpass API呼び出しフロー

1. **アクティブなルール取得**: `is_active = true` のルールをDBから取得
2. **ループ処理**: 各ルールに対して以下を実行
   - ルールに紐づくキーワードを取得
   - 各キーワードでconnpass API v2を呼び出し（1秒間隔、APIキー必須）
   - 取得したイベント情報を `events_cache` に保存
   - 通知条件判定（新規公開・申込開始・残席わずか・締切前）
   - 条件に合致すればDiscord通知送信
   - 送信履歴を `notifications` テーブルに記録（UNIQUE制約で重複防止）
3. **データクリーンアップ**:
   - `events_cache`: 2週間以上前のデータを削除
   - `notifications`: 2週間以上前のデータを削除
   - `important_logs`: 3ヶ月以上前のデータを削除

### 🚨 エラーハンドリング

- **connpass API エラー**: `important_logs` に記録して次のキーワードへスキップ
- **Discord送信エラー**: `important_logs` に記録して次のイベントへスキップ
- **DB接続エラー**: スケジューラ全体を停止し、`scheduler_status.last_error` に記録

---

## 11. エラーログ管理

### 📝 important_logs テーブルの活用

**記録対象のイベント**:

| ログレベル | イベントタイプ | 記録タイミング |
|-----------|---------------|---------------|
| **ERROR** | `connpass_api_error` | connpass API呼び出し失敗 |
| **ERROR** | `discord_send_failed` | Discord通知送信失敗 |
| **ERROR** | `auth_error` | OAuth2認証エラー |
| **ERROR** | `database_error` | DB操作エラー |
| **WARNING** | `rate_limit_warning` | APIレート制限警告 |
| **INFO** | `scheduler_start` | スケジューラ起動 |
| **INFO** | `scheduler_complete` | スケジューラ正常完了 |

### 🗑️ ログ保持期間

- **保持期間**: 3ヶ月
- **自動削除**: スケジューラ実行時に3ヶ月以上前のログまたは、保存可能数を超えた場合古いログを削除
- **理由**: エラー調査に十分な期間を確保しつつ、DBサイズを抑制

---

## 12. ディレクトリ構成

### 📁 プロジェクト全体

```
connpass-discord-notifier/
├── backend/                    # Goバックエンド
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go         # APIサーバー起動
│   │   ├── bot/
│   │   │   └── main.go         # Discord Bot起動
│   │   └── scheduler/
│   │       └── main.go         # スケジューラ起動
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go       # 環境変数読み込み
│   │   ├── database/
│   │   │   └── db.go           # DB接続
│   │   ├── handlers/
│   │   │   ├── auth.go         # 認証ハンドラ
│   │   │   ├── rules.go        # ルールCRUD
│   │   │   ├── status.go       # ステータスAPI
│   │   │   └── logs.go         # ログAPI
│   │   ├── services/
│   │   │   ├── connpass.go     # connpass API クライアント
│   │   │   ├── discord.go      # Discord通知・チャンネル作成
│   │   │   ├── notifier.go     # 通知判定ロジック
│   │   │   ├── scheduler.go    # スケジューラ本体
│   │   │   └── logger.go       # ログ記録サービス
│   │   ├── repository/
│   │   │   ├── user.go
│   │   │   ├── rule.go
│   │   │   ├── event.go
│   │   │   ├── notification.go
│   │   │   └── log.go
│   │   └── models/
│   │       ├── user.go
│   │       ├── rule.go
│   │       ├── event.go
│   │       └── log.go
│   ├── migrations/             # SQLマイグレーション
│   │   ├── 001_create_users.sql
│   │   ├── 002_create_guilds.sql
│   │   ├── 003_create_rules.sql
│   │   ├── 004_create_rule_keywords.sql
│   │   ├── 005_create_rule_notify_types.sql
│   │   ├── 006_create_events_cache.sql
│   │   ├── 007_create_notifications.sql
│   │   ├── 008_create_important_logs.sql
│   │   └── 009_create_scheduler_status.sql
│   ├── tests/
│   │   ├── unit/
│   │   └── integration/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   └── .env.example
│
├── frontend/                   # Next.jsフロントエンド
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx            # ランディングページ
│   │   ├── login/
│   │   │   └── page.tsx
│   │   ├── dashboard/
│   │   │   └── page.tsx
│   │   ├── rules/
│   │   │   ├── page.tsx        # ルール一覧
│   │   │   ├── new/
│   │   │   │   └── page.tsx    # ルール作成
│   │   │   └── [id]/
│   │   │       └── edit/
│   │   │           └── page.tsx # ルール編集
│   │   ├── status/
│   │   │   └── page.tsx        # ステータス監視画面
│   │   └── api/
│   │       └── auth/
│   │           └── callback/
│   │               └── route.ts
│   ├── components/
│   │   ├── ui/                 # shadcn/ui コンポーネント
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── select.tsx
│   │   │   ├── input.tsx
│   │   │   └── ...
│   │   └── features/
│   │       ├── rule-form.tsx
│   │       ├── channel-selector.tsx
│   │       ├── test-notification.tsx
│   │       └── log-viewer.tsx
│   ├── lib/
│   │   ├── api.ts              # Go APIクライアント
│   │   ├── auth.ts             # JWT管理
│   │   └── types.ts            # TypeScript型定義
│   ├── public/
│   ├── Dockerfile
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   ├── next.config.js
│   └── .env.local.example
│
├── docs/                       # ドキュメント
│   ├── api.md                  # API仕様書
│   └── deployment.md           # デプロイ手順
│
├── docker-compose.yml
└── README.md                   # このファイル
```

---

## 13. デプロイ構成

### 🚂 Railway デプロイ

#### サービス構成

| サービス名 | 起動コマンド | ポート | 実行モード |
|-----------|-------------|--------|-----------|
| **api-service** | `/api` | 8080 | 常駐 |
| **bot-service** | `/bot` | - | 常駐（WebSocket） |
| **scheduler-service** | `/scheduler` | - | Cron (0,30 * * * *) |
| **postgres** | (managed) | 5432 | 常駐 |

---

## 14. 開発環境セットアップ

### 🖥️ 必要な環境

- **Go**: 1.25.2 linux/amd64
- **Node.js**
- **Docker**
- **Docker Compose**

### 🔑 connpass APIキーの取得

**connpass API v2を利用するには、APIキーの取得が必須です:**

1. **connpassにログイン**: https://connpass.com/login/
2. **API設定ページにアクセス**: https://connpass.com/about/api/
3. **APIキーを発行**: 画面の指示に従ってAPIキーを生成
4. **環境変数に設定**: 取得したAPIキーを `.env` ファイルに追加

```bash
CONNPASS_API_KEY=your_actual_api_key_here
```

**注意事項**:
- API v1は2025年末に廃止予定です
- API v2はすべてのリクエストに `X-API-Key` ヘッダーが必須
- レート制限: 1秒間に1リクエストまで

---

## 15. セキュリティ・運用

### 🔒 セキュリティ対策

| 項目 | 対策内容 |
|------|---------|
| **通信の暗号化** | すべてHTTPS通信（Railway・Vercel自動対応） |
| **認証** | Discord OAuth2 + JWT（HttpOnly Cookie） |
| **パスワード管理** | Discordに委譲（自前でパスワード管理しない） |
| **環境変数** | Railway・Vercelの環境変数機能で管理 |
| **CORS** | フロントエンドのドメインのみ許可 |
| **SQLインジェクション対策** | プリペアドステートメント使用 |
| **XSS対策** | Next.jsのデフォルトエスケープ機能 |
| **API認証** | すべてのAPIエンドポイントでJWT検証 |

### 📊 APIレート制限対策

**connpass API v2**:
- **APIキー認証**: 必須（`X-API-Key` ヘッダー）
- **レート制限**: 1秒間に1リクエストまで
- リクエスト間隔: 1秒
- 429エラー時: 60秒待機してリトライ

**Discord API**:
- メッセージ送信間隔: 1秒
- レート制限ヘッダーを監視

### 🗄️ データバックアップ

**Railway PostgreSQL**:
- 自動バックアップ: 7日間保持
- 手動バックアップ: `pg_dump`コマンド

---

## 16. コスト試算

### 💰 月額コスト（Hobby プラン想定）

| サービス | プラン | 月額 | 備考 |
|---------|-------|------|------|
| **Railway** | Hobby | $5 | $5分のリソース込み |
| **Railway 超過分** | 使用量課金 | $5〜10 | CPU・RAM使用量 |
| **Railway PostgreSQL** | 1GB | $5 | データベース |
| **Vercel** | Hobby | $0 | 無料枠で十分 |
| **Discord Bot** | - | $0 | 無料 |
| **connpass API** | - | $0 | 無料 |

**月額合計**: **$15〜20/月**

### 📊 リソース使用量の目安

**CPU使用量**:
- API Server (常駐): 0.5 vCPU × 24h × 30日 = 360 CPU時間
- Bot (常駐): 0.2 vCPU × 24h × 30日 = 144 CPU時間
- Scheduler (30分ごと): 0.1 vCPU × 2分 × 48回/日 × 30日 = 4.8 CPU時間
- **合計**: 約509 CPU時間

**Railway課金レート** (目安):
- CPU: 約$0.02/vCPU時間
- RAM: 約$0.000231/GB時間

**試算**:
- CPU: 509時間 × $0.02 = 約$10
- RAM: (0.5GB + 0.25GB) × 24h × 30日 × $0.000231 = 約$1.25
- **合計**: 約$11〜12の超過料金

---

## 17. 今後の拡張予定

### 🚀 Phase 2

- [ ] Discord Slashコマンド対応 (`/check`, `/rules`)
- [ ] 通知メッセージのカスタマイズ（色・絵文字・テンプレート）
- [ ] マルチサーバー対応の強化
- [ ] UIからのBot招待フロー

### 🚀 Phase 3

- [ ] イベントのお気に入り機能
- [ ] 通知の一時停止機能
- [ ] 統計ダッシュボード（通知数グラフなど）
- [ ] Webhook対応（Slack・LINE通知）

---

## サポート

### 問題が発生した場合

1. **ステータス画面を確認**: `/status`でシステム状態を確認
2. **ログを確認**: `important_logs`テーブルを参照
3. **Railwayログを確認**: Railway管理画面で実行ログを確認

### 開発リソース

- **connpass API v2仕様**: https://connpass.com/about/api/v2/
- **connpass API v2 OpenAPI定義**: https://connpass.com/about/api/v2/openapi.json
- **Discord API仕様**: https://discord.com/developers/docs
- **discordgo ドキュメント**: https://pkg.go.dev/github.com/bwmarrin/discordgo
- **Next.js ドキュメント**: https://nextjs.org/docs
- **Railway ドキュメント**: https://docs.railway.com

---

## 📝 変更履歴

- **2025-11-11**: connpass API v2対応
  - API v2への移行（APIキー認証必須）
  - レート制限を2秒→1秒に変更
  - エンドポイント変更（`/api/v1/event/` → `/api/v2/events/`）
  - レスポンス構造の変更（`event_id` → `id`, `series` → `group`）
  - 環境変数 `CONNPASS_API_KEY` を追加

- **2025-11-05**: 初版作成
  - Go 1.25.2対応
  - important_logsテーブル追加
  - スケジューラ30分間隔に変更
  - Railwayコスト情報更新（Hobby $5プランで試験運用）

---

**作成日**: 2025-11-05
**最終更新**: 2025-11-11
**バージョン**: 1.1
**作成者**: [rikut0904]