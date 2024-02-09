package handlers

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
	"github.com/zYoma/gophermart/internal/utils"
)

func (h *HandlerService) CreateOrder(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, models.Error("empty body"))
		return
	}

	orderNumber := string(body)
	if !utils.CheckLuhn(orderNumber) {
		logger.Log.Error("номер заказа не валидный")
		w.WriteHeader(http.StatusUnprocessableEntity)
		render.JSON(w, r, models.Error("not valid order number"))
		return
	}

	userID, err := getUserFromRequest(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		render.JSON(w, r, models.Error("Unauthorized"))
		return
	}

	err = h.provider.CreateOrder(r.Context(), orderNumber, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrCreatedByOtherUser) {
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, models.Error("order created by other user"))
			return
		}
		if errors.Is(err, postgres.ErrOrderAlredyExist) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		logger.Log.Error("ошибка при запросе в БД", zap.Error(err))
		render.JSON(w, r, models.Error("error create order"))
		return
	}

	go func() {
		select {
		case h.orderChan <- orderNumber:
			// Успешно отправлено
		case <-time.After(time.Second * 5):
			// Обработка таймаута, возможно запись в лог
			logger.Log.Error("timeout when sending to orderChan")
		}
	}()

	w.WriteHeader(http.StatusAccepted)

}