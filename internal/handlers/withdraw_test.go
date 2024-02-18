package handlers

import (
	"bytes"
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
	service := New(providerMock, cfg)
	r := service.GetRouter()
	srv := httptest.NewServer(r)
	defer srv.Close()

	testCases := []struct {
		name          string
		method        string
		body          any
		expectedCode  int
		sum           float64
		expectedError error
	}{
		{
			name:          "успешный кейс",
			method:        http.MethodPost,
			body:          models.OrderSum{Sum: 100, Order: "2377225624"},
			expectedCode:  http.StatusOK,
			sum:           100,
			expectedError: nil,
		},
		{
			name:          "не валидный номер заказа",
			method:        http.MethodPost,
			body:          models.OrderSum{Sum: 200, Order: "12345"},
			expectedCode:  http.StatusUnprocessableEntity,
			sum:           200,
			expectedError: nil,
		},
		{
			name:          "недостаточно средств",
			method:        http.MethodPost,
			body:          models.OrderSum{Sum: 1000, Order: "2377225624"},
			expectedCode:  http.StatusPaymentRequired,
			sum:           1000,
			expectedError: postgres.ErrFewPoints,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настройка поведения моков
			providerMock.On("Withdrow", mock.Anything, tc.sum, mock.Anything, mock.Anything).Return(tc.expectedError)

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
