package handlers

import (
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
)

func TestHandlerService_GetBalance(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)
	token, _ := jwt.BuildJWTString("user", cfg.TokenSecret)
	token2, _ := jwt.BuildJWTString("jack", cfg.TokenSecret)

	// Настройка поведения моков
	mockBalance := models.Balance{
		Current:   400,
		Withdrawn: 43,
	}

	service := New(providerMock, cfg)
	r := service.GetRouter()
	srv := httptest.NewServer(r)
	defer srv.Close()

	testCases := []struct {
		name         string
		method       string
		expectedCode int
		token        string
		expectedBody models.Balance
		user         string
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
			token:        token2,
			expectedBody: mockBalance,
			user:         "jack",
		},
		{
			name:         "нет баланса",
			method:       http.MethodGet,
			expectedCode: http.StatusOK,
			token:        token,
			expectedBody: models.Balance{},
			user:         "user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			providerMock.On("GetUserBalance", mock.Anything, tc.user).Return(tc.expectedBody, nil)
			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/balance", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
			require.NoError(t, err)

			// Выполнение запроса
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			var response models.Balance
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)
			assert.Equal(t, response, tc.expectedBody)

		})
	}
}
