package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"github.com/zYoma/gophermart/internal/auth/hash"
	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
	"go.uber.org/zap"
)

func (h *HandlerService) Registration(w http.ResponseWriter, r *http.Request) {

	var credentials models.Credantials

	w.Header().Set("Content-Type", "application/json")
	if err := decodeAndValidateBody(w, r, &credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passHash, err := hash.HashPassword(credentials.Password)
	if err != nil {
		logger.Log.Error("не удалось получить хеш пароля", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		render.JSON(w, r, models.Error("error create user"))
		return
	}

	err = h.provider.CreateUser(r.Context(), credentials.Login, passHash)
	if err != nil {
		if errors.Is(err, postgres.ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, models.Error("user already exist"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при записи в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error create user"))
		return
	}

	token, err := jwt.BuildJWTString(credentials.Login, h.cfg.TokenSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка генерации токена", zap.Error(err))
		render.JSON(w, r, models.Error("error create user"))
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
	response := models.AccessToken{Token: token, TokenType: "Bearer"}

	render.JSON(w, r, &response)

}

func decodeAndValidateBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	err := render.DecodeJSON(r.Body, dst)
	if errors.Is(err, io.EOF) {
		logger.Log.Error("request body is empty")
		render.JSON(w, r, models.Error("empty request"))
		return err
	}
	if err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		render.JSON(w, r, models.Error("failed to decode request"))
		return err
	}

	validate := validator.New()
	if err := validate.Struct(dst); err != nil {
		validateErr := err.(validator.ValidationErrors)
		logger.Log.Error("request validate error", zap.Error(err))
		render.JSON(w, r, models.ValidationError(validateErr))
		return err
	}

	return nil
}
