package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/integrations/loyalty"
	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/storage"
	"go.uber.org/zap"
)

// с определённым интервалом проверяет начисления в системе лояльности для заказов в статусе REGISTERED
func UpdateOrdersStatus(cfg *config.Config, wg *sync.WaitGroup, provider storage.Provider, orderChan chan string) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(cfg.CheckOrderInterval) * time.Second)
	defer ticker.Stop()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case order := <-orderChan:
			// загружен новый заказ, делаем запрос в систему лояльности
			orderProccessed(ctx, order, provider, cfg)

		case <-ticker.C:
			// сработал таймер
			registeredOrders := getOrders(ctx, provider)
			startProccessed(ctx, registeredOrders, provider, cfg)
		case <-ctx.Done():
			return
		}
	}
}

func startProccessed(ctx context.Context, orders []string, provider storage.Provider, cfg *config.Config) {
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		// Избегаем проблемы захвата переменной в замыкании, копируя значение в локальную переменную цикла
		order := order
		go func() {
			orderProccessed(ctx, order, provider, cfg)
		}()
	}

}

func getOrders(ctx context.Context, provider storage.Provider) []string {
	orders, err := provider.GetRegisteresOrders(ctx)
	if err != nil {
		logger.Log.Error("cannot get orders", zap.Error(err))
		return []string{}
	}
	return orders
}

func orderProccessed(ctx context.Context, order string, provider storage.Provider, cfg *config.Config) {
	orderResp, err := loyalty.GetPointsByOrder(fmt.Sprintf("%s/api/orders/%s", cfg.AcrualURL, order))
	if err != nil {
		if errors.Is(err, loyalty.ErrNotFound) {
			orderResp = &loyalty.OrderResponse{Order: order, Status: "PROCESSING"}
		} else {
			logger.Log.Error("не удалось получить данные по заказу", zap.Error(err))
			return
		}
	}

	// обмновляем данные по заказу и пополняем баланс пользователя
	errDB := provider.UpdateOrderAndAccrualPoints(ctx, orderResp)
	if errDB != nil {
		logger.Log.Error("не удалось обновить заказ", zap.Error(err))
		return
	}

	logger.Log.Sugar().Infof("заказ %s обработан. Записан статус: %s", order, orderResp.Status)
}
