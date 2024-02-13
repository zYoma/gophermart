package loyalty

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/zYoma/gophermart/internal/logger"
	"go.uber.org/zap"
)

type OrderStatus string

const (
	StatusRegistered OrderStatus = "REGISTERED"
	StatusInvalid    OrderStatus = "INVALID"
	StatusProcessing OrderStatus = "PROCESSING"
	StatusProcessed  OrderStatus = "PROCESSED"
)

type OrderResponse struct {
	Order   string      `json:"order"`
	Status  OrderStatus `json:"status"`
	Accrual *float64    `json:"accrual,omitempty"`
}

var ErrRequest = errors.New("request to loyalty")
var ErrReadBody = errors.New("read body")
var ErrUnmarshal = errors.New("unmarshal response")
var ErrStatus = errors.New("bad order status")
var ErrStatusCode = errors.New("not success status")

// isValid проверяет, является ли статус заказа допустимым.
func (s OrderStatus) isValid() bool {
	switch s {
	case StatusRegistered, StatusInvalid, StatusProcessing, StatusProcessed:
		return true
	default:
		return false
	}
}

func GetPointsByOrder(url string) (*OrderResponse, error) {
	resp, err := http.Get(url)
	if err != nil {
		logger.Log.Error("ошибка при выполнении запроса", zap.Error(err))
		return nil, ErrRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Log.Sugar().Infof("сервер вернул статус-код: %d, url: %s", resp.StatusCode, url)
		return nil, ErrStatusCode
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("ошибка при чтении тела ответа", zap.Error(err))
		return nil, ErrReadBody
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		logger.Log.Error("ошибка при десериализации ответа", zap.Error(err))
		return nil, ErrUnmarshal
	}

	if !orderResp.Status.isValid() {
		logger.Log.Sugar().Infof("недопустимый статус заказа: %s, url: %s", orderResp.Status, url)
		return nil, ErrStatus
	}

	return &orderResp, nil
}
