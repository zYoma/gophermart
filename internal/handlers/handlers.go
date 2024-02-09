package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/storage"
)

type HandlerService struct {
	provider  storage.Provider
	cfg       *config.Config
	orderChan chan string
}

func New(provider storage.Provider, cfg *config.Config) *HandlerService {
	return &HandlerService{provider: provider, cfg: cfg, orderChan: make(chan string, 1024)}
}

func (h *HandlerService) GetRouter() chi.Router {

	r := chi.NewRouter()

	r.Use(handlerLogger)
	r.Use(h.jwtAuthMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", h.Registration)
		r.Post("/api/user/login", h.Login)
		r.Post("/api/user/orders", h.CreateOrder)
		r.Get("/api/user/orders", h.GetOrders)
		r.Get("/api/user/balance", h.GetBalance)
		r.Post("/api/user/balance/withdraw", h.WithdrowPoints)
		r.Get("/api/user/withdrawals", h.GetWithdrawals)
	})

	return r
}
