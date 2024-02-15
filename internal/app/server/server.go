package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/zYoma/gophermart/internal/app/tasks"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/handlers"
	"github.com/zYoma/gophermart/internal/storage"
)

type HTTPServer struct {
	server *http.Server
	wg     *sync.WaitGroup
}

func New(
	ctx context.Context,
	provider storage.Provider,
	cfg *config.Config,
) *HTTPServer {

	// создаем сервис обработчик
	service := handlers.New(provider, cfg)

	// запускаем горутину для обработки заказов
	var wg sync.WaitGroup
	wg.Add(1)
	go tasks.UpdateOrdersStatus(ctx, cfg, &wg, provider)

	// получаем роутер
	router := service.GetRouter()

	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: router,
	}
	return &HTTPServer{
		server: server,
		wg:     &wg,
	}
}

func (a *HTTPServer) Run() error {
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (a *HTTPServer) Shutdown(ctx context.Context) error {
	a.wg.Wait()
	return a.server.Shutdown(ctx)
}
