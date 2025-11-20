# API 仕様概要

## 認証

- Discord OAuth2でログインする。
- 認証後に発行されたJWTは`session` Cookieとして返却される。
- APIを呼び出す際はCookie送信、または`Authorization: Bearer <token>`ヘッダーを付与する。

## エンドポイント一覧

### POST `/api/auth/callback`
- 認可コードを受け取り、Discordトークン交換・ユーザー登録を実施。
- 成功時: `200 OK`
  ```json
  {
    "token": "<JWT>",
    "user": { ... },
    "guilds": [ ... ]
  }
  ```

### GET `/api/me/guilds`
- 認証中ユーザーの管理可能なギルドを返す。
- 成功時: `200 OK`
  ```json
  [
    {
      "id": 1,
      "guildId": "123",
      "guildName": "Sample Guild",
      "canManage": true
    }
  ]
  ```

### GET `/api/guilds/:guildId/channels`
- Botが取得できるテキストチャンネルを返す。
- 成功時: `200 OK`
  ```json
  {
    "channels": [{ "id": "456", "name": "general" }]
  }
  ```

### GET `/api/rules?guild_id=xxxx`
- 指定ギルドの通知ルールを取得。

### POST `/api/rules`
- ルール新規作成。
  ```json
  {
    "guildId": "123",
    "channelId": "456",
    "name": "Go勉強会",
    "keywords": ["Go", "Golang"],
    "notifyTypes": ["open", "almost_full"],
    "isActive": true
  }
  ```

### GET `/api/rules/:id`
- ルール詳細取得。

### PUT `/api/rules/:id`
- ルール更新。

### DELETE `/api/rules/:id`
- ルール削除。

### POST `/api/rules/:id/test`
- 指定ルールの設定チャンネルにテスト通知を送信。

### GET `/api/status`
- スケジューラの最新状態。

### GET `/api/logs?limit=20`
- 重要ログを新しい順に取得。

## エラーレスポンス
- 共通フォーマット: `{"message": "エラーメッセージ"}`
- HTTPステータスコードを併せて確認すること。
