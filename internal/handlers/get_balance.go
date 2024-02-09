package handlers

import (
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
)

func (h *HandlerService) GetBalance(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserFromRequest(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, models.Error("Unauthorized"))
		return
	}

	balance, err := h.provider.GetUserBalance(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при запросе в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error get balance"))
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, models.Balance(balance))
}
