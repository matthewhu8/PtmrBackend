package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fb "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	mockfb "github.com/hankimmy/PtmrBackend/pkg/firebase/mock"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	mockwk "github.com/hankimmy/PtmrBackend/pkg/worker/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUserFirebase(t *testing.T) {
	mockToken := fmt.Sprintf(`{"uid": "mock-uid", "role": "%s"}`, db.RoleEmployer)
	encodedToken := base64.StdEncoding.EncodeToString([]byte(mockToken))
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request)
		body          gin.H
		buildStubs    func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"name":  "Test User",
				"email": "test@example.com",
				"role":  db.RoleEmployer,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(&fb.Token{UID: "mock-uid"}, nil)

				auth.EXPECT().SetUserRole(gomock.Any(), "mock-uid", string(db.RoleEmployer)).
					Times(1).
					Return(nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), "test@example.com", gomock.Eq(string(db.RoleEmployer))).
					Times(1).
					Return("http://example.com/verify?token=abc123", nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Name:             "Test User",
					Email:            "test@example.com",
					VerificationLink: "http://example.com/verify?token=abc123",
				}

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"name":  "Test User",
				"email": "test@example.com",
				"role":  "user_role",
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().
					VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(nil, errors.New("unauthorized"))

				auth.EXPECT().
					SetUserRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				auth.EXPECT().
					GenerateEmailVerificationLink(gomock.Any(), gomock.Any(), gomock.Eq(string(db.RoleEmployer))).
					Times(0)

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalServerErrorOnRoleSetting",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", "valid-firebase-id-token")
			},
			body: gin.H{
				"name":  "Test User",
				"email": "test@example.com",
				"role":  "user_role",
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&fb.Token{UID: "some-uid"}, nil)

				auth.EXPECT().SetUserRole(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerErrorOnEmailVerificationLink",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", "valid-firebase-id-token")
			},
			body: gin.H{
				"name":  "Test User",
				"email": "test@example.com",
				"role":  db.RoleCandidate,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&fb.Token{UID: "some-uid"}, nil)

				auth.EXPECT().SetUserRole(gomock.Any(), "some-uid", string(db.RoleCandidate)).
					Times(1).
					Return(nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), "test@example.com", gomock.Eq(string(db.RoleCandidate))).
					Times(1).
					Return("", errors.New("internal server error"))

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			authCtrl := gomock.NewController(t)
			defer authCtrl.Finish()
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()

			auth := mockfb.NewMockAuthClientFirebase(authCtrl)
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			tc.buildStubs(auth, taskDistributor)

			server := newTestServer(t, nil, taskDistributor, nil, auth, nil)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)
			request := createNewRequest(t, http.MethodPost, "/users/", data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestResendEmailVerification(t *testing.T) {
	mockToken := fmt.Sprintf(`{"uid": "mock-uid", "role": "%s"}`, db.RoleEmployer)
	cMockToken := fmt.Sprintf(`{"uid": "mock-uid", "role": "%s"}`, db.RoleCandidate)
	encodedToken := base64.StdEncoding.EncodeToString([]byte(mockToken))
	cEncodedToken := base64.StdEncoding.EncodeToString([]byte(cMockToken))
	testCases := []struct {
		name             string
		setupAuth        func(t *testing.T, request *http.Request)
		body             gin.H
		buildStubs       func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor)
		setupRateLimiter func(rateLimiter *RateLimiter) // For setting up rate limiting conditions
		checkResponse    func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"email": "test@example.com",
				"role":  db.RoleEmployer,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(&fb.Token{UID: "mock-uid"}, nil)

				auth.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
					Times(1).
					Return(&fb.UserRecord{
						UserInfo: &fb.UserInfo{
							DisplayName: "Test User",
						},
						EmailVerified: false,
					}, nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), "test@example.com", gomock.Eq(string(db.RoleEmployer))).
					Times(1).
					Return("http://example.com/verify?token=abc123", nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Name:             "Test User",
					Email:            "test@example.com",
					VerificationLink: "http://example.com/verify?token=abc123",
				}

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			setupRateLimiter: func(rateLimiter *RateLimiter) {
				// No rate limiting in this case
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "Unauthorized",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"email": "test@example.com",
				"role":  "user_role",
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(nil, errors.New("unauthorized"))

				auth.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(0)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupRateLimiter: func(rateLimiter *RateLimiter) {
				// No need to set rate limiter since it's unauthorized
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "EmailAlreadyVerified",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"email": "test@example.com",
				"role":  db.RoleEmployer,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(&fb.Token{UID: "mock-uid"}, nil)

				auth.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
					Times(1).
					Return(&fb.UserRecord{EmailVerified: true}, nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupRateLimiter: func(rateLimiter *RateLimiter) {
				// No need for rate limiter since email is already verified
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "RateLimiting",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleEmployer))
			},
			body: gin.H{
				"email": "test@example.com",
				"role":  db.RoleEmployer,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), encodedToken).
					Times(1).
					Return(&fb.Token{UID: "mock-uid"}, nil)

				auth.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
					Times(1).
					Return(&fb.UserRecord{EmailVerified: false}, nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupRateLimiter: func(rateLimiter *RateLimiter) {
				// Simulate that the email was sent less than a minute ago
				rateLimiter.SetLastSent("mock-uid", time.Now().Add(-30*time.Second))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusTooManyRequests, recorder.Code)
			},
		},
		{
			name: "InternalServerErrorOnEmailVerificationLink",
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, "Bearer", string(db.RoleCandidate))
			},
			body: gin.H{
				"email": "test@example.com",
				"role":  db.RoleCandidate,
			},
			buildStubs: func(auth *mockfb.MockAuthClientFirebase, taskDistributor *mockwk.MockTaskDistributor) {
				auth.EXPECT().VerifyIDToken(gomock.Any(), cEncodedToken).
					Times(1).
					Return(&fb.Token{UID: "mock-uid"}, nil)

				auth.EXPECT().GetUserByEmail(gomock.Any(), "test@example.com").
					Times(1).
					Return(&fb.UserRecord{EmailVerified: false}, nil)

				auth.EXPECT().GenerateEmailVerificationLink(gomock.Any(), "test@example.com", string(db.RoleCandidate)).
					Times(1).
					Return("", errors.New("internal server error"))

				taskDistributor.EXPECT().DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			setupRateLimiter: func(rateLimiter *RateLimiter) {
				// No rate limiter required for internal server error test case
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			authCtrl := gomock.NewController(t)
			defer authCtrl.Finish()
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()

			auth := mockfb.NewMockAuthClientFirebase(authCtrl)
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)

			// Initialize real RateLimiter
			rateLimiter := NewRateLimiter()

			tc.buildStubs(auth, taskDistributor)
			tc.setupRateLimiter(rateLimiter)

			server := newTestServer(t, nil, taskDistributor, nil, auth, rateLimiter)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)
			request := createNewRequest(t, http.MethodPost, "/resend-verification-email", data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
