# 環境変数設定ガイド

このドキュメントでは、開発・ステージング・本番で利用する環境変数ファイルの作成手順と必須項目を整理します。**秘密情報はリポジトリにコミットしない**よう注意してください。

---

## 1. バックエンド (`backend/.env`)

### 作成手順

```bash
cd backend
cp .env.example .env
```

コピー後、`.env` 内の各値を環境に合わせて編集します。Railway などのホスティングを利用する場合は、同じ値を管理画面にも登録してください。

> **補足**  
> ローカル開発では `.env.example` をベースに `.env` を作成し、Railway へデプロイする際は `backend/.env.railway.example` を参考にして環境変数を登録します（Railway 側はダッシュボードで直接設定する運用を想定）。

### 主な環境変数

| 変数名 | 必須 | 説明 | 例 | 備考 |
|--------|------|------|----|------|
| `PORT` | 任意 | API の待ち受けポート | `8080` | Railway ではデフォルトポートを利用 |
| `DATABASE_URL` | 必須 | PostgreSQL 接続文字列 | `postgres://user:pass@host:5432/db?sslmode=disable` | 開発では `docker-compose.yml` の値を利用 |
| `JWT_SECRET` | 必須 | セッション JWT の署名鍵 | `change-me` | 32 文字以上のランダム値を推奨 |
| `DISCORD_CLIENT_ID` | 必須 | Discord OAuth2 クライアント ID | `123456789012345678` | Discord Developer Portal で取得 |
| `DISCORD_CLIENT_SECRET` | 必須 | Discord OAuth2 クライアントシークレット | `xxxxxxxxxxxxxxxx` | 同上。漏洩注意 |
| `DISCORD_REDIRECT_URI` | 必須 | Discord リダイレクト URL | `http://localhost:3000/login` | Vercel では `https://your-app.vercel.app/login` 等に変更 |
| `DISCORD_BOT_TOKEN` | 必須 | Discord Bot トークン | `Bot <token>` | `Bot ` は自動付与されるため値は `<token>` 部分のみ |
| `DISCORD_PUBLIC_KEY` | 任意 | Slash コマンドなどで利用 | （未使用） | 使わない場合は空のままで問題なし |
| `CONNPASS_BASE_URL` | 任意 | connpass API エンドポイント | `https://connpass.com/api/v2/events/` | 変更不要 |
| `CONNPASS_API_KEY` | 必須 | connpass API キー | `your-api-key` | connpass で API キーを取得 |
| `CONNPASS_REQUEST_INTERVAL` | 任意 | connpass 呼び出し間隔 | `1s` | レート制限に合わせて調整 |
| `NOTIFICATION_DEFAULT_THRESHOLD` | 任意 | 「残席わずか」判定の既定閾値 | `80` | ルール側で上書き可能 |
| `SCHEDULER_POLL_INTERVAL` | 任意 | スケジューラ実行間隔 | `30m` | Railway の Cron 設定と整合させる |
| `SESSION_MODE` | 任意 | セッション有効期間モード | `production` | develop: 1分, production: 3ヶ月 |

### 取り扱いの注意
- `.env` はローカル開発専用です。本番ではプラットフォームの環境変数管理機能を使ってください。
- `JWT_SECRET` や `DISCORD_CLIENT_SECRET` などの秘密値は、パスワードマネージャや Vault に保管します。

---

## 2. フロントエンド (`frontend/.env.local`)

### 作成手順

```bash
cd frontend
cp .env.local.example .env.local
```

Next.js は、`.env.local` を自動読み込みします（ローカル限定）。Vercel でデプロイする場合は、同じ値を「Project Settings » Environment Variables」に登録してください。

### 主な環境変数

| 変数名 | 必須 | 説明 | 例 | 備考 |
|--------|------|------|----|------|
| `NEXT_PUBLIC_API_BASE_URL` | 必須 | バックエンド API のベース URL | `http://localhost:8080/api` | 本番では Railway の公開 URL を指定 |
| `NEXT_PUBLIC_DISCORD_CLIENT_ID` | 必須 | Discord OAuth2 クライアント ID | `123456789012345678` | バックエンド設定と揃える |
| `NEXT_PUBLIC_DISCORD_REDIRECT_URI` | 必須 | Discord リダイレクト URL | `http://localhost:3000/login` | Vercel URL に合わせて変更 |
| `NEXT_PUBLIC_DISCORD_OAUTH_URL` | 任意 | Discord OAuth エンドポイント | `https://discord.com/oauth2/authorize` | 基本的に変更不要 |

> `NEXT_PUBLIC_` プレフィックス付きの変数はクライアント側へ公開されるため、秘密情報は設定しないでください。

---

## 3. Docker Compose 利用時

`docker-compose.yml` は `DISCORD_CLIENT_ID` や `DISCORD_BOT_TOKEN` などをホスト環境から参照します。ローカルで `docker compose up` を実行する前に、以下のいずれかを行ってください。

1. シェル上で直接エクスポートする
   ```bash
   export DISCORD_CLIENT_ID=xxxx
   export DISCORD_CLIENT_SECRET=yyyy
   export DISCORD_BOT_TOKEN=zzzz
   docker compose up
   ```
2. プロジェクト直下に `.env`（Docker 用）を用意し、同様のキーを記載する。

---

## 4. 推奨ワークフロー

1. `.env.example` / `.env.local.example` をそれぞれコピーして環境別ファイルを作成
2. 必須項目に実際の値を入力
3. ローカルで `go run ./cmd/api`、`npm run dev`、`docker compose up` などを実行して接続確認
4. ホスティング環境（Railway / Vercel 等）へ同じ値を登録  
   - Railway: `backend/.env.railway.example` をベースに Variables を設定（`.env` ファイルのアップロードは不要）

### よくあるトラブル
- **Discord ログインに失敗**: `DISCORD_REDIRECT_URI` と Discord Developer Portal の設定が一致しているか確認
- **Bot のチャンネル取得ができない**: `DISCORD_BOT_TOKEN` が正しいか、ボットがサーバーに参加しているかをチェック
- **API 401 エラー**: `JWT_SECRET` がフロントとバックエンドで一致しているか確認

---

## 5. セキュリティベストプラクティス
- `.gitignore` に `.env`, `.env.local` が含まれていることを確認
- 機密情報は Issue / Pull Request / ログ に貼らない
- 本番鍵は定期的にローテーションを行い、アクセス権限を最小限に保つ

---

これらの手順に従って環境変数を設定すれば、ローカル開発から本番運用まで同一構成で動作確認が可能になります。
