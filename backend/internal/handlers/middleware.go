package handlers

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"connpass-requirement/internal/config"
)

type contextKey string

const (
	contextKeyUserID      contextKey = "userID"
	contextKeyDiscordUser contextKey = "discordUserID"
)

// JWTMiddleware はJWTを検証し、ユーザー情報をContextに格納する。
func JWTMiddleware(cfg config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := extractToken(c)
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token claims")
			}

			userIDFloat, ok := claims["sub"].(float64)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid user id")
			}

			discordUserID, _ := claims["discord_user_id"].(string)

			c.Set(string(contextKeyUserID), int64(userIDFloat))
			c.Set(string(contextKeyDiscordUser), discordUserID)

			return next(c)
		}
	}
}

func extractToken(c echo.Context) string {
	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	auth := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// MustUserID はコンテキストからユーザーIDを取り出す。
func MustUserID(c echo.Context) int64 {
	if v, ok := c.Get(string(contextKeyUserID)).(int64); ok {
		return v
	}
	return 0
}

// MustDiscordUserID はDiscordユーザーIDを取り出す。
func MustDiscordUserID(c echo.Context) string {
	if v, ok := c.Get(string(contextKeyDiscordUser)).(string); ok {
		return v
	}
	return ""
}
