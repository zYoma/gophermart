package storage

import (
	"context"

	"github.com/zYoma/gophermart/internal/integrations/loyalty"
	"github.com/zYoma/gophermart/internal/models"
)

type StorageProvider interface {
	Provider
}

type Provider interface {
	Init() error
	CreateUser(ctx context.Context, login string, password string) error
	GetPasswordHash(ctx context.Context, login string) (string, error)
	CreateOrder(ctx context.Context, number string, login string) error
	GetRegisteresOrders(ctx context.Context) ([]string, error)
	UpdateOrderAndAccrualPoints(ctx context.Context, orderData *loyalty.OrderResponse) error
	GetUserOrders(ctx context.Context, user_login string) ([]models.Order, error)
	GetUserBalance(ctx context.Context, user_login string) (models.Balance, error)
	Withdrow(ctx context.Context, sum float64, user_login string, order string) error
	GetUserWithdrawals(ctx context.Context, user_login string) ([]models.Withdrawn, error)
}
