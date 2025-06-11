package handler

import (
	"bytes"
	"errors"
	"net/http"
	"testing"
	"todoai/internal/models"
	"todoai/internal/repository"
	"todoai/internal/service"
	"todoai/internal/service/mocks"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandler_signUp(t *testing.T) {
	type mockBehavior func(s *mocks.MockAuthService, user models.FirstAuth)

	testTable := []struct {
		name           string
		user           models.FirstAuth
		input          string
		mockBehavior   mockBehavior
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			user: models.FirstAuth{
				Email:    "lip@example.com",
				Login:    "root",
				Password: "password123",
			},
			input: `{"email": "lip@example.com", "login": "root", "password": "password123"}`,
			mockBehavior: func(s *mocks.MockAuthService, user models.FirstAuth) {
				s.EXPECT().SignUp(gomock.Any(), gomock.Eq(&user)).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"user created successfully"}`,
		},
		{
			name: "user exists",
			user: models.FirstAuth{
				Email:    "lip@example.com",
				Login:    "root",
				Password: "password123",
			},
			input: `{"email": "lip@example.com", "login": "root", "password": "password123"}`,
			mockBehavior: func(s *mocks.MockAuthService, user models.FirstAuth) {
				s.EXPECT().SignUp(gomock.Any(), gomock.Eq(&user)).Return(repository.ErrUserExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"user already exists"}`,
		},
		{
			name:           "invalid response",
			user:           models.FirstAuth{},
			input:          `{"email": "lip@example.com", "login": "root", "password": "pass"}`,
			mockBehavior:   func(s *mocks.MockAuthService, user models.FirstAuth) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid response"}`,
		},
		{
			name: "failed to authenticate user",
			user: models.FirstAuth{
				Email:    "lip@example.com",
				Login:    "root",
				Password: "password123",
			},
			input: `{"email": "lip@example.com", "login": "root", "password": "password123"}`,
			mockBehavior: func(s *mocks.MockAuthService, user models.FirstAuth) {
				s.EXPECT().SignUp(gomock.Any(), gomock.Eq(&user)).Return(errors.New("failed to create user"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to create user"}`,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authServiceMock := mocks.NewMockAuthService(ctrl)
			tt.mockBehavior(authServiceMock, tt.user)

			h := &handler{
				service: &service.Service{
					Auth: authServiceMock,
				},
			}
			r := gin.New()
			r.POST("/sign-up", h.signUp)

			w := httptest.NewRecorder()

			req := httptest.NewRequest("POST", "/sign-up", bytes.NewBufferString(tt.input))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}

}

func TestHandler_signIn(t *testing.T) {
	type mockBehavior func(s *mocks.MockAuthService, user models.SecondAuth)

	validUser := models.SecondAuth{
		Login:    "testuser",
		Password: "password123",
		DeviceID: "device-1",
	}

	validInput := `{"login":"testuser","password":"password123","device_id":"device-1"}`

	tokens := models.Tokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	testTable := []struct {
		name           string
		input          string
		user           models.SecondAuth
		mockBehavior   mockBehavior
		expectedStatus int
		expectedBody   string
	}{
		{
			name:  "OK",
			input: validInput,
			user:  validUser,
			mockBehavior: func(s *mocks.MockAuthService, user models.SecondAuth) {
				s.EXPECT().SignIn(gomock.Any(), &user).
					Return(tokens, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"access_token":"access-token","refresh_token":"refresh-token"}`,
		},
		{
			name:  "Invalid JSON",
			input: `{"login":"testuser", "password":123}`,
			user:  models.SecondAuth{},
			mockBehavior: func(s *mocks.MockAuthService, user models.SecondAuth) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid response"}`,
		},
		{
			name:  "Auth error",
			input: validInput,
			user:  validUser,
			mockBehavior: func(s *mocks.MockAuthService, user models.SecondAuth) {
				s.EXPECT().SignIn(gomock.Any(), &user).
					Return(models.Tokens{}, errors.New("failed to authenticate user"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to authenticate user"}`,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authService := mocks.NewMockAuthService(ctrl)
			tt.mockBehavior(authService, tt.user)

			h := &handler{
				service: &service.Service{
					Auth: authService,
				},
			}
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/sign-in", h.signIn)

			req := httptest.NewRequest("POST", "/sign-in", bytes.NewBufferString(tt.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}

}

func TestHandler_verify(t *testing.T) {
	type mockBehavior func(s *mocks.MockAuthService, code string)

	testCases := []struct {
		name           string
		inputBody      string
		code           string
		mockBehavior   mockBehavior
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "success",
			inputBody: `{"code":"123456"}`,
			code:      "123456",
			mockBehavior: func(s *mocks.MockAuthService, code string) {
				s.EXPECT().Verify(gomock.Any(), code).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"user verified successfully"}`,
		},
		{
			name:      "invalid json",
			inputBody: `{"code":123}`,
			code:      "",
			mockBehavior: func(s *mocks.MockAuthService, code string) {
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"invalid response"}`,
		},
		{
			name:      "user not found",
			inputBody: `{"code":"123456"}`,
			code:      "123456",
			mockBehavior: func(s *mocks.MockAuthService, code string) {
				s.EXPECT().Verify(gomock.Any(), code).Return(repository.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"invalid response"}`,
		},
		{
			name:      "internal error",
			inputBody: `{"code":"123456"}`,
			code:      "123456",
			mockBehavior: func(s *mocks.MockAuthService, code string) {
				s.EXPECT().Verify(gomock.Any(), code).Return(errors.New("some error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to verify user"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authService := mocks.NewMockAuthService(ctrl)
			tt.mockBehavior(authService, tt.code)

			h := &handler{
				service: &service.Service{
					Auth: authService,
				},
			}
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.POST("/verify", h.verify)

			req := httptest.NewRequest("POST", "/verify", bytes.NewBufferString(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			require.Equal(t, tt.expectedStatus, w.Code)
			require.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}
