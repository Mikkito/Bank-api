package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"bank-api/internal/config"
	"bank-api/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey = contextKey("userID")
)

// JWTMiddleware check JWT and add userID in context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWT.Secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID
		userID := claims.Subject
		if userID == "" {
			http.Error(w, "Invalid token subject", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Get userID
func GetUserID(ctx context.Context) (int64, error) {
	idStr, ok := ctx.Value(UserIDKey).(string)
	if !ok || idStr == "" {
		return 0, errors.New("user ID not found in context")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	return id, nil
}

func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
			return
		}

		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			utils.RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "jwt secret not configured"})
			return
		}

		// Разбор токена
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
			return
		}

		// Извлекаем user_id
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			utils.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid user ID in token"})
			return
		}
		userID := int64(userIDFloat)

		// Вставляем userID в контекст
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
