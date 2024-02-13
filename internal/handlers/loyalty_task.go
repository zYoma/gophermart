package handlers

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zYoma/gophermart/internal/integrations/loyalty"
	"github.com/zYoma/gophermart/internal/logger"
	"go.uber.org/zap"
)

// с определённым интервалом проверяет начисления в системе лояльности для заказов в статусе REGISTERED
func (h *HandlerService) UpdateOrdersStatus(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(h.cfg.CheckOrderInterval) * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case order := <-h.orderChan:
			// загружен новый заказ, делаем запрос в систему лояльности
			h.orderProccessed(ctx, order)

		case <-ticker.C:
			// сработал таймер
			registeredOrders := h.getOrders(ctx)
			h.startProccessed(ctx, registeredOrders)
		case <-ctx.Done():
			return
		}
	}
}

func (h *HandlerService) startProccessed(ctx context.Context, orders []string) {
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		// Избегаем проблемы захвата переменной в замыкании, копируя значение в локальную переменную цикла
		order := order
		go func() {
			h.orderProccessed(ctx, order)
		}()
	}

}

func (h *HandlerService) getOrders(ctx context.Context) []string {
	orders, err := h.provider.GetRegisteresOrders(ctx)
	if err != nil {
		logger.Log.Error("cannot get orders", zap.Error(err))
		return []string{}
	}
	return orders
}

func (h *HandlerService) orderProccessed(ctx context.Context, order string) {
	orderResp, err := loyalty.GetPointsByOrder(fmt.Sprintf("%s/api/orders/%s", h.cfg.AcrualURL, order))
	if err != nil {
		if errors.Is(err, loyalty.ErrNotFound) {
			orderResp = &loyalty.OrderResponse{Order: order, Status: "INVALID"}
		} else {
			logger.Log.Error("не удалось получить данные по заказу", zap.Error(err))
			return
		}
	}

	// обмновляем данные по заказу и пополняем баланс пользователя
	errDB := h.provider.UpdateOrderAndAccrualPoints(ctx, orderResp)
	if errDB != nil {
		logger.Log.Error("не удалось обновить заказ", zap.Error(err))
		return
	}

	logger.Log.Sugar().Infof("заказ %s обработан. Записан статус: %s", order, orderResp.Status)
}
