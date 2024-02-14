package loyalty

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

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

var (
	ErrRequest    = errors.New("request to loyalty")
	ErrReadBody   = errors.New("read body")
	ErrUnmarshal  = errors.New("unmarshal response")
	ErrStatus     = errors.New("bad order status")
	ErrStatusCode = errors.New("not success status")
	ErrNotFound   = errors.New("order not found")
)

// isValid проверяет, является ли статус заказа допустимым.
func (s OrderStatus) isValid() bool {
	switch s {
	case StatusRegistered, StatusInvalid, StatusProcessing, StatusProcessed:
		return true
	default:
		return false
	}
}

func GetPointsByOrder(url string, pauseChan chan bool) (*OrderResponse, error) {
	for {
		select {
		case <-pauseChan: // При получении сигнала о паузе, ожидаем следующий сигнал для продолжения
			<-pauseChan
		default: // Продолжаем выполнение, если пауза не активирована
			resp, err := http.Get(url)
			if err != nil {
				logger.Log.Error("ошибка при выполнении запроса", zap.Error(err))
				return nil, ErrRequest
			}
			// повторяем запрос при статусе 429
			if resp.StatusCode == http.StatusTooManyRequests {
				retryAfter := resp.Header.Get("Retry-After")
				delaySeconds, err := strconv.Atoi(retryAfter)
				if err != nil {
					logger.Log.Sugar().Infof("ошибка при чтении заголовка Retry-After", err)
					resp.Body.Close()
					return nil, ErrStatusCode
				}
				logger.Log.Sugar().Infof("Получен статус 429, повтор запроса через %d секунд\n", delaySeconds)
				resp.Body.Close()
				broadcastPause(pauseChan) // Отправляем сигнал о паузе всем горутинам
				time.Sleep(time.Duration(delaySeconds) * time.Second)
				broadcastResume(pauseChan) // Отправляем сигнал о возобновлении работы всем горутинам
				continue
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// если заказ не найден
				if resp.StatusCode == http.StatusNoContent {
					logger.Log.Sugar().Infof("заказ не найден, url: %s", url)
					return nil, ErrNotFound
				}
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
	}
}

func broadcastPause(pauseChan chan bool) {
	// Отправляем сигнал о паузе, не блокируя, если канал уже заполнен
	select {
	case pauseChan <- true:
	default:
	}
}

func broadcastResume(pauseChan chan bool) {
	// Отправляем сигнал о возобновлении, не блокируя, если канал уже заполнен
	select {
	case pauseChan <- false:
	default:
	}
}
