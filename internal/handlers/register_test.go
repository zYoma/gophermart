package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zYoma/gophermart/internal/config"
	"github.com/zYoma/gophermart/internal/mocks"
	"github.com/zYoma/gophermart/internal/models"
	"github.com/zYoma/gophermart/internal/storage/postgres"
)

func TestHandlerService_Registration(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)

	// Настройка поведения моков
	providerMock.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, login string, password string) error {
			if login == "jack" {
				return postgres.ErrConflict
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
		expectedBody string
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodPost,
			body:         models.Credantials{Login: "user", Password: "password"},
			expectedCode: http.StatusOK,
			expectedBody: "Bearer",
		},
		{
			name:         "пустое тело запроса",
			method:       http.MethodPost,
			body:         nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "имя уже занято",
			method:       http.MethodPost,
			body:         models.Credantials{Login: "jack", Password: "password"},
			expectedCode: http.StatusConflict,
			expectedBody: "",
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
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/register", &buf)
			require.NoError(t, err)

			// Выполнение запроса
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Проверки
			assert.Equal(t, tc.expectedCode, resp.StatusCode)
			if tc.expectedBody != "" {
				var response models.AccessToken
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response.TokenType, tc.expectedBody)
			}
		})
	}
}

func GetMockConfig() *config.Config {
	return &config.Config{
		RunAddr:     ":8081",
		AcrualURL:   "http://localhost:8080",
		LogLevel:    "info",
		TokenSecret: "test",
	}
}
