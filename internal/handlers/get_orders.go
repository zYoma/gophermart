package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
)

func (h *HandlerService) GetOrders(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	userID, err := getUserFromRequest(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, models.Error("Unauthorized"))
		return
	}

	orders, err := h.provider.GetUserOrders(r.Context(), userID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrdersNotFound) {
			w.WriteHeader(http.StatusNoContent)
			render.JSON(w, r, models.Error("orders not found"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при запросе в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error get orders"))
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, models.Orders(orders))
}
