package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connpass-requirement/internal/config"
)

// OAuthService はDiscord OAuth2フローを扱う。
type OAuthService struct {
	cfg    config.Config
	client *http.Client
}

func NewOAuthService(cfg config.Config) *OAuthService {
	return &OAuthService{
		cfg:    cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ExchangeCode は認可コードをアクセストークンへ交換する。
func (o *OAuthService) ExchangeCode(ctx context.Context, code string) (*OAuthTokenResponse, error) {
	endpoint := "https://discord.com/api/v10/oauth2/token"
	data := url.Values{}
	data.Set("client_id", o.cfg.DiscordClientID)
	data.Set("client_secret", o.cfg.DiscordClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", o.cfg.DiscordRedirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("discord token exchange failed: %s", string(body))
	}

	var token OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// FetchIdentity はユーザー情報と所属ギルドを取得する。
func (o *OAuthService) FetchIdentity(ctx context.Context, token string) (*DiscordUser, []DiscordGuild, error) {
	user, err := o.getUser(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	guilds, err := o.getGuilds(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	return user, guilds, nil
}

func (o *OAuthService) getUser(ctx context.Context, token string) (*DiscordUser, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/users/@me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("discord user fetch failed: %s", string(body))
	}

	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (o *OAuthService) getGuilds(ctx context.Context, token string) ([]DiscordGuild, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/users/@me/guilds", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("discord guild fetch failed: %s", string(body))
	}

	var guilds []DiscordGuild
	if err := json.NewDecoder(resp.Body).Decode(&guilds); err != nil {
		return nil, err
	}
	return guilds, nil
}

// OAuthTokenResponse はDiscordのトークンレスポンス。
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// DiscordUser はDiscordユーザー情報。
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}

// DiscordGuild はユーザーが所属するギルド情報。
type DiscordGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Permissions string `json:"permissions"`
}
