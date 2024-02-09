package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/mocks"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
)

func TestHandlerService_Withdraw(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)
	token, _ := jwt.BuildJWTString("user", cfg.TokenSecret)

	// Настройка поведения моков
	providerMock.On("Withdrow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, sum float64, userLogin string, order string) error {
			if sum == 1000 {
				return postgres.ErrFewPoints
			}
			return nil
		})
	service := New(providerMock, cfg)
	r := service.GetRouter()
	srv := httptest.NewServer(r)
	defer srv.Close()

	testCases := []struct {
		name         string
		method       string
		body         any
		expectedCode int
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodPost,
			body:         models.OrderSum{Sum: 100, Order: "2377225624"},
			expectedCode: http.StatusOK,
		},
		{
			name:         "не валидный номер заказа",
			method:       http.MethodPost,
			body:         models.OrderSum{Sum: 200, Order: "12345"},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "недостаточно средств",
			method:       http.MethodPost,
			body:         models.OrderSum{Sum: 1000, Order: "2377225624"},
			expectedCode: http.StatusPaymentRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Подготовка тела запроса
			var buf bytes.Buffer
			if tc.body != nil {
				err := json.NewEncoder(&buf).Encode(tc.body)
				require.NoError(t, err)
			}

			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/balance/withdraw", &buf)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			require.NoError(t, err)

			// Выполнение запроса
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tc.expectedCode, resp.StatusCode)

		})
	}
}
