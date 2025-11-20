# デプロイ手順 (概要)

## Railway へのデプロイ

1. リポジトリをRailwayに接続し、Backend用サービスを作成する。
2. `backend/Dockerfile`をビルド対象に設定し、環境変数を登録する。
   - `DATABASE_URL`
   - `JWT_SECRET`
   - `DISCORD_CLIENT_ID`
   - `DISCORD_CLIENT_SECRET`
   - `DISCORD_REDIRECT_URI`
   - `DISCORD_BOT_TOKEN`
   - `CONNPASS_BASE_URL`
3. 同一リポジトリからBot用サービス (`/bot`) と Scheduler用ジョブ (`/scheduler`) を作成。
4. PostgreSQLサービスを追加し、`DATABASE_URL`を共有する。
5. `railway run go run cmd/scheduler/main.go` で初回テスト実行。

## Vercel へのデプロイ

1. `frontend` ディレクトリをVercelに接続する。
2. 環境変数
   - `NEXT_PUBLIC_API_BASE_URL`
   - `NEXT_PUBLIC_DISCORD_CLIENT_ID`
   - `NEXT_PUBLIC_DISCORD_REDIRECT_URI`
3. ビルドコマンド: `npm run build`
4. 出力ディレクトリ: `.next`

## Discord Bot 設定

1. Discord Developer Portalでアプリケーション/ボットを作成。
2. BotトークンをRailwayに設定。
3. OAuth2 URL generatorからBot権限 (`MANAGE_CHANNELS`, `SEND_MESSAGES`) を選択し招待。

## マイグレーション

- `backend/migrations` ディレクトリのSQLを `goose` や `tern` などで実行。
- ローカルでは `docker-compose up` で自動的にPostgreSQLが起動する。

## ローカル開発

```bash
docker-compose up --build
```

- http://localhost:3000 でフロントエンド
- http://localhost:8080/api でAPI
