package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type PublicKeyResponse struct {
	PublicKey string `json:"public_key"`
	ExpiresAt string `json:"expires_at"`
}

func JWTMiddleware(
	publicKeyURL string,
	cacheTTL int,
	log *logger.Logger,
) gin.HandlerFunc {
	var cachedKey *rsa.PublicKey
	var keyExpiresAt time.Time

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		publicKey, err := getPublicKey(ctx, publicKeyURL, cacheTTL, &cachedKey, &keyExpiresAt, log)
		if err != nil {
			log.Error("Failed to get public key", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication service unavailable"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil || !token.Valid {
			log.Warn("Invalid token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Set("email", claims["email"])

		c.Next()
	}
}

func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found in token"})
			c.Abort()
			return
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getPublicKey(
	ctx context.Context,
	url string,
	cacheTTL int,
	cachedKey **rsa.PublicKey,
	keyExpiresAt *time.Time,
	log *logger.Logger,
) (*rsa.PublicKey, error) {
	// Проверяем кэш
	if *cachedKey != nil && time.Now().Before(*keyExpiresAt) {
		log.Debug("Using cached public key")
		return *cachedKey, nil
	}

	// Запрашиваем от Auth сервиса
	log.Info("Fetching public key from auth service", zap.String("url", url))

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var keyResp PublicKeyResponse
	if err := json.Unmarshal(body, &keyResp); err != nil {
		return nil, fmt.Errorf("failed to parse public key response: %w", err)
	}

	// Парсим RSA public key
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyResp.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA key: %w", err)
	}

	// Кэшируем
	*cachedKey = publicKey
	*keyExpiresAt = time.Now().Add(time.Duration(cacheTTL) * time.Second)

	log.Info("Public key cached", zap.Time("expires_at", *keyExpiresAt))
	return publicKey, nil
}
