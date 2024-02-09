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

func TestHandlerService_GetWithdrawals(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)
	token, _ := jwt.BuildJWTString("user", cfg.TokenSecret)
	token2, _ := jwt.BuildJWTString("jack", cfg.TokenSecret)

	// Настройка поведения моков
	mockWithdrawals := []models.Withdrawn{
		{
			Order:       "123",
			Sum:         500,
			ProccesedAt: time.Now(),
		},
		{
			Order:       "456",
			Sum:         600,
			ProccesedAt: time.Now().Add(-48 * time.Hour),
		},
	}

	providerMock.On("GetUserWithdrawals", mock.Anything, mock.Anything).Return(
		mockWithdrawals, func(ctx context.Context, userLogin string) error {
			if userLogin == "user" {
				return postgres.ErrWithdrawalsNotFound
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
		expectedBody models.Withdrawals
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
			token:        token2,
			expectedBody: models.Withdrawals(mockWithdrawals),
		},
		{
			name:         "нет списаний",
			method:       http.MethodGet,
			expectedCode: http.StatusNoContent,
			token:        token,
			expectedBody: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/withdrawals", nil)
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
