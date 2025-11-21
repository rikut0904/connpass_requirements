package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/config"
	"connpass-requirement/internal/models"
	"connpass-requirement/internal/repository"
	"connpass-requirement/internal/services"
)

// AuthHandler はDiscord OAuth2コールバックを処理する。
type AuthHandler struct {
	cfg    config.Config
	oauth  *services.OAuthService
	users  *repository.UserRepository
	logger *services.LoggerService
}

func NewAuthHandler(
	cfg config.Config,
	oauth *services.OAuthService,
	users *repository.UserRepository,
	logger *services.LoggerService,
) *AuthHandler {
	return &AuthHandler{cfg: cfg, oauth: oauth, users: users, logger: logger}
}

// RegisterAuthRoutes は認証系ルートを登録する。
func RegisterAuthRoutes(g *echo.Group, handler *AuthHandler) {
	g.POST("/auth/callback", handler.HandleCallback)
}

// RegisterAuthRoutesWithMiddleware は認証が必要なルートを登録する。
func RegisterAuthRoutesWithMiddleware(g *echo.Group, handler *AuthHandler) {
	g.GET("/auth/me", handler.HandleMe)
	g.POST("/auth/logout", handler.HandleLogout)
}

// HandleLogout はログアウト処理を行う。
func (h *AuthHandler) HandleLogout(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, map[string]string{"message": "logged out"})
}

// HandleMe は現在のユーザー情報を返す。
func (h *AuthHandler) HandleMe(c echo.Context) error {
	userID := MustUserID(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	user, err := h.users.FindByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	guilds, err := h.users.GetGuildPermissions(c.Request().Context(), user.ID)
	if err != nil {
		guilds = []models.GuildPermission{}
	}

	return c.JSON(http.StatusOK, authCallbackResponse{
		User:   *user,
		Guilds: guilds,
	})
}

type authCallbackRequest struct {
	Code string `json:"code"`
}

type authCallbackResponse struct {
	Token  string                   `json:"token"`
	User   models.User              `json:"user"`
	Guilds []models.GuildPermission `json:"guilds"`
}

// HandleCallback はDiscord OAuth2コールバック。
func (h *AuthHandler) HandleCallback(c echo.Context) error {
	var req authCallbackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if strings.TrimSpace(req.Code) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code is required")
	}

	token, err := h.oauth.ExchangeCode(c.Request().Context(), req.Code)
	if err != nil {
		h.logger.Error(c.Request().Context(), "auth_error", "Discordトークン取得に失敗", err)
		return echo.NewHTTPError(http.StatusBadGateway, "failed to exchange discord token")
	}

	identity, guilds, err := h.oauth.FetchIdentity(c.Request().Context(), token.AccessToken)
	if err != nil {
		h.logger.Error(c.Request().Context(), "auth_error", "Discordユーザー情報取得に失敗", err)
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch discord identity")
	}

	user := models.User{
		DiscordUserID:   identity.ID,
		DiscordUsername: identity.Username,
		AvatarURL:       buildAvatarURL(identity.ID, identity.Avatar),
		AccessToken:     token.AccessToken,
		RefreshToken:    token.RefreshToken,
		TokenExpiresAt:  time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}

	if err := h.users.Upsert(c.Request().Context(), &user); err != nil {
		h.logger.Error(c.Request().Context(), "database_error", "ユーザー情報の保存に失敗", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save user")
	}

	guildPerms := make([]models.GuildPermission, 0, len(guilds))
	for _, guild := range guilds {
		permValue, _ := strconv.ParseInt(guild.Permissions, 10, 64)
		guildPerms = append(guildPerms, models.GuildPermission{
			GuildID:       guild.ID,
			GuildName:     guild.Name,
			IconURL:       buildGuildIconURL(guild.ID, guild.Icon),
			Permissions:   permValue,
			CanManage:     hasPermission(permValue, 0x8) || hasPermission(permValue, 0x20),
			CanManageRole: hasPermission(permValue, 0x20),
		})
	}

	if err := h.users.SaveGuildPermissions(c.Request().Context(), user.ID, guildPerms); err != nil {
		h.logger.Error(c.Request().Context(), "database_error", "ギルド権限保存に失敗", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save guild permissions")
	}

	jwtToken, err := h.issueJWT(c.Request().Context(), user)
	if err != nil {
		h.logger.Error(c.Request().Context(), "auth_error", "JWT発行に失敗", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to issue token")
	}

	cookie := &http.Cookie{
		Name:     "session",
		Value:    jwtToken,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(h.cfg.SessionDuration),
	}
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, authCallbackResponse{
		Token:  jwtToken,
		User:   user,
		Guilds: guildPerms,
	})
}

func (h *AuthHandler) issueJWT(ctx context.Context, user models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":             user.ID,
		"discord_user_id": user.DiscordUserID,
		"exp":             time.Now().Add(h.cfg.SessionDuration).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.cfg.JWTSecret))
}

func buildAvatarURL(userID, avatar string) string {
	if avatar == "" {
		return ""
	}
	return "https://cdn.discordapp.com/avatars/" + userID + "/" + avatar + ".png"
}

func buildGuildIconURL(guildID, icon string) string {
	if icon == "" {
		return ""
	}
	return "https://cdn.discordapp.com/icons/" + guildID + "/" + icon + ".png"
}

func hasPermission(value int64, bit int64) bool {
	return value&bit == bit
}
