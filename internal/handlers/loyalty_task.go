package handlers

import (
	"context"
	"fmt"
	"os"
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

	// будем проверять заказы в статусе REGISTERED раз в минуту
	ticker := time.NewTicker(60 * time.Second)

	// Канал для перехвата сигналов завершения работы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case order := <-h.orderChan:
			// загружен новый заказ, делаем запрос в систему лояльности
			h.orderProccessed(order)

		case <-ticker.C:
			// сработал таймер
			registeredOrders := h.getOrders()
			h.startProccessed(registeredOrders)
		case <-sigChan:
			return
		}
	}
}

func (h *HandlerService) startProccessed(orders []string) {
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		// Избегаем проблемы захвата переменной в замыкании, копируя значение в локальную переменную цикла
		order := order
		go func() {
			h.orderProccessed(order)
		}()
	}

}

func (h *HandlerService) getOrders() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orders, err := h.provider.GetRegisteresOrders(ctx)
	if err != nil {
		logger.Log.Error("cannot get orders", zap.Error(err))
		return []string{}
	}
	return orders
}

func (h *HandlerService) orderProccessed(order string) {
	orderResp, err := loyalty.GetPointsByOrder(fmt.Sprintf("%s/api/orders/%s", h.cfg.AcrualURL, order))
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// обмновляем данные по заказу и пополняем баланс пользователя
	errDB := h.provider.UpdateOrderAndAccrualPoints(ctx, orderResp)
	if errDB != nil {
		logger.Log.Error("не удалось обновить заказ", zap.Error(err))
	}

	logger.Log.Sugar().Infof("заказ %s обработан. Записан статус: %s", order, orderResp.Status)
}
