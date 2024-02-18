package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zYoma/gophermart/internal/mocks"
	"github.com/zYoma/gophermart/internal/models"
)

func TestHandlerService_Login(t *testing.T) {
	cfg := GetMockConfig()

	providerMock := new(mocks.StorageProvider)
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
		user         string
		passHash     string
	}{
		{
			name:         "успешный кейс",
			method:       http.MethodPost,
			body:         models.Credantials{Login: "user", Password: "1234"},
			expectedCode: http.StatusOK,
			expectedBody: "Bearer",
			user:         "user",
			passHash:     "$2a$10$dheHgk3mKFTybDiYQ6RmfeLTeBZMOcrNTqA1DMU5uxNJi0dth34wm",
		},
		{
			name:         "пустое тело запроса",
			method:       http.MethodPost,
			body:         nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			user:         "",
			passHash:     "",
		},
		{
			name:         "не верный пароль",
			method:       http.MethodPost,
			body:         models.Credantials{Login: "jack", Password: "password"},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "",
			user:         "jack",
			passHash:     "232",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Настройка поведения моков
			providerMock.On("GetPasswordHash", mock.Anything, tc.user).Return(tc.passHash, nil)
			// Подготовка тела запроса
			var buf bytes.Buffer
			if tc.body != nil {
				err := json.NewEncoder(&buf).Encode(tc.body)
				require.NoError(t, err)
			}

			// Создание запроса
			req, err := http.NewRequest(tc.method, srv.URL+"/api/user/login", &buf)
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
