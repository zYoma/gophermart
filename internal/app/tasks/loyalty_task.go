package tasks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/integrations/loyalty"
	"github.com/zYoma/gophermart/internal/logger"
	"github.com/zYoma/gophermart/internal/storage"
	"go.uber.org/zap"
)

type TaskService struct {
	provider storage.Provider
	cfg      *config.Config
	wg       *sync.WaitGroup
}

func New(provider storage.Provider, cfg *config.Config, wg *sync.WaitGroup) *TaskService {
	return &TaskService{provider: provider, cfg: cfg, wg: wg}
}

// с определённым интервалом проверяет начисления в системе лояльности для заказов с не конечными статусами
func (t *TaskService) UpdateOrdersStatus(ctx context.Context) {
	defer t.wg.Done()

	ticker := time.NewTicker(time.Duration(t.cfg.CheckOrderInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// сработал таймер
			registeredOrders := t.getOrders(ctx)
			t.startProccessed(ctx, registeredOrders)
		case <-ctx.Done():
			return
		}
	}
}

func (t *TaskService) startProccessed(ctx context.Context, orders []string) {
	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		// Избегаем проблемы захвата переменной в замыкании, копируя значение в локальную переменную цикла
		order := order
		go func() {
			OrderProccessed(ctx, order, t.provider, t.cfg)
		}()
	}

}

func (t *TaskService) getOrders(ctx context.Context) []string {
	orders, err := t.provider.GetRegisteresOrders(ctx)
	if err != nil {
		logger.Log.Error("cannot get orders", zap.Error(err))
		return nil
	}
	return orders
}

func OrderProccessed(ctx context.Context, order string, provider storage.Provider, cfg *config.Config) {
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
