package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
	"github.com/zYoma/gophermart/internal/utils"
	"go.uber.org/zap"
)

func (h *HandlerService) WithdrowPoints(w http.ResponseWriter, r *http.Request) {

	var orderSum models.OrderSum

	w.Header().Set("Content-Type", "application/json")
	if err := decodeAndValidateBody(w, r, &orderSum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := getUserFromRequest(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, models.Error("Unauthorized"))
		return
	}

	if !utils.CheckLuhn(orderSum.Order) {
		logger.Log.Error("номер заказа не валидный")
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, models.Error("not valid order number"))
		return
	}

	err = h.provider.Withdrow(r.Context(), orderSum.Sum, userID, orderSum.Order)
	if err != nil {
		if errors.Is(err, postgres.ErrFewPoints) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			render.JSON(w, r, models.Error("there are not enough points on balance"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при запросе в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error create order"))
		return
	}

	w.WriteHeader(http.StatusOK)

}
