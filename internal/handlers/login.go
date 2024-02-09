package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"

	"github.com/zYoma/gophermart/internal/auth/hash"
	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"go.uber.org/zap"
)

func (h *HandlerService) Login(w http.ResponseWriter, r *http.Request) {

	var credentials models.Credantials

	w.Header().Set("Content-Type", "application/json")
	if err := decodeAndValidateBody(w, r, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passwordHash, err := h.provider.GetPasswordHash(r.Context(), credentials.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при запросе в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error find user"))
		return
	}

	if !hash.CheckPassword(passwordHash, credentials.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Log.Error("неверная пара логин/пароль")
		render.JSON(w, r, models.Error("wrong credentials"))
		return
	}

	token, err := jwt.BuildJWTString(credentials.Login, h.cfg.TokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка генерации токена", zap.Error(err))
		render.JSON(w, r, models.Error("error create user"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	response := models.AccessToken{Token: token, TokenType: "Bearer"}

	render.JSON(w, r, &response)

}
