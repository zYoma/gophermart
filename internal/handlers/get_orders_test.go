package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/mocks"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
)

func TestHandlerService_GetOrders(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)
	token, _ := jwt.BuildJWTString("user", cfg.TokenSecret)
	token2, _ := jwt.BuildJWTString("jack", cfg.TokenSecret)

	// Настройка поведения моков
	accrualValue1 := 400.0
	accrualValue2 := 500.0
	mockOrders := []models.Order{
		{
			Number:     "123",
			Status:     "PROCESSED",
			Accrual:    &accrualValue1,
			UploadedAt: time.Now(),
		},
		{
			Number:     "456",
			Status:     "PROCESSED",
			Accrual:    &accrualValue2,
			UploadedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			Number:     "789",
			Status:     "PROCESSING",
			UploadedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	providerMock.On("GetUserOrders", mock.Anything, mock.Anything).Return(
		mockOrders, func(ctx context.Context, userLogin string) error {
			if userLogin == "user" {
				return postgres.ErrOrdersNotFound
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
		expectedCode int
		token        string
		expectedBody models.Orders
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
			token:        token2,
			expectedBody: models.Orders(mockOrders),
		},
		{
			name:         "нет заказов",
			method:       http.MethodGet,
			expectedCode: http.StatusNoContent,
			token:        token,
			expectedBody: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/orders", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			require.NoError(t, err)

			// Выполнение запроса
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tc.expectedCode, resp.StatusCode)
			if tc.expectedBody != nil {
				var response models.Orders
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, len(response), len(tc.expectedBody))
			}
		})
	}
}
