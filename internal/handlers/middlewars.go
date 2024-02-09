package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/logger"
	"go.uber.org/zap"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

var noAuthRequired = []string{
	"/api/user/register",
	"/api/user/login",
}
var ErrGetUserFromRequest = errors.New("faild get user")

func handlerLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создание ResponseRecorder для перехвата ответа
		recorder := &responseRecorder{w, 0, 0}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		logger.Log.Info("handlerLogger",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
			zap.Int("status", recorder.status),
			zap.Int64("size", recorder.size),
		)

	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int64
}

// Переопределение WriteHeader для сохранения реального статуса ответа
func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Переопределение Write для сохранения размера ответа
func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += int64(size)
	return size, err
}

func (h *HandlerService) jwtAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := h.cfg.TokenSecret

		if !pathRequiresAuth(r.RequestURI) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID := jwt.GetUserID(token, secret)
		if userID == "" {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		// Передаем идентификатор пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromRequest(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", ErrGetUserFromRequest
	}
	return userID, nil
}

// Функция проверки пути на наличие в списке исключений.
func pathRequiresAuth(path string) bool {
	for _, p := range noAuthRequired {
		if p == path {
			return false
		}
	}
	return true
}
