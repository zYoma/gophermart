package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zYoma/gophermart/internal/auth/jwt"
	"github.com/zYoma/gophermart/internal/mocks"
	"github.com/zYoma/gophermart/internal/storage/postgres"
)

func TestHandlerService_CreateOrder(t *testing.T) {
	cfg := GetMockConfig()
	providerMock := new(mocks.StorageProvider)
	token, _ := jwt.BuildJWTString("user", cfg.TokenSecret)

	// Настройка поведения моков
	providerMock.On("CreateOrder", mock.Anything, mock.Anything, mock.Anything).Return(
		func(ctx context.Context, number string, login string) error {
			if number == "4111111111111111" {
				return postgres.ErrOrderAlredyExist
			}
			if number == "2377225624" {
				return postgres.ErrCreatedByOtherUser
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
		body         string
		expectedCode int
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodPost,
			body:         "79927398713",
			expectedCode: http.StatusAccepted,
		},
		{
			name:         "был загружен ранее",
			method:       http.MethodPost,
			body:         "4111111111111111",
			expectedCode: http.StatusOK,
		},
		{
			name:         "был загружен другим юзером",
			method:       http.MethodPost,
			body:         "2377225624",
			expectedCode: http.StatusConflict,
		},
		{
			name:         "не валидный номер",
			method:       http.MethodPost,
			body:         "237722562444",
			expectedCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			body := bytes.NewBufferString(tc.body)
			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/orders", body)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			req.Header.Set("Content-Type", "text/plain")
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
