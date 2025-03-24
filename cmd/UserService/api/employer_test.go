package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/hankimmy/PtmrBackend/pkg/db/mock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestGetEmployerAPI(t *testing.T) {
	user, _ := randomUser(t)
	employer := randomEmployer(user.Username)

	testCases := []struct {
		name          string
		employerID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return(employer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployer(t, recorder.Body, employer)
			},
		},
		{
			name:       "UnauthorizedUser",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "unauthorized_user", db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return(employer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "NoAuthorization",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "NotFound",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return(db.Employer{}, db.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			employerID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil, nil, nil, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/employers/%d", tc.employerID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateEmployerAPI(t *testing.T) {
	user, _ := randomUser(t)
	employer := randomEmployer(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"business_name":        employer.BusinessName,
				"business_email":       employer.BusinessEmail,
				"business_phone":       employer.BusinessPhone,
				"location":             employer.Location,
				"industry":             employer.Industry,
				"profile_photo":        employer.ProfilePhoto,
				"business_description": employer.BusinessDescription,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateEmployerParams{
					Username:            employer.Username,
					BusinessName:        employer.BusinessName,
					BusinessEmail:       employer.BusinessEmail,
					BusinessPhone:       employer.BusinessPhone,
					Location:            employer.Location,
					Industry:            employer.Industry,
					ProfilePhoto:        employer.ProfilePhoto,
					BusinessDescription: employer.BusinessDescription,
				}
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(employer, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployer(t, recorder.Body, employer)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"business_name":        employer.BusinessName,
				"business_email":       employer.BusinessEmail,
				"business_phone":       employer.BusinessPhone,
				"location":             employer.Location,
				"industry":             employer.Industry,
				"profile_photo":        employer.ProfilePhoto,
				"business_description": employer.BusinessDescription,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"business_name":        employer.BusinessName,
				"business_email":       employer.BusinessEmail,
				"business_phone":       employer.BusinessPhone,
				"location":             employer.Location,
				"industry":             employer.Industry,
				"profile_photo":        employer.ProfilePhoto,
				"business_description": employer.BusinessDescription,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateEmployer(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Employer{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil, nil, nil, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/employers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListEmployersAPI(t *testing.T) {
	user, _ := randomUser(t)
	n := 5
	employers := make([]db.Employer, n)
	for i := 0; i < n; i++ {
		employers[i] = randomEmployer(user.Username)
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.ListEmployersParams{
					Username: user.Username,
					Limit:    int32(n),
					Offset:   0,
				}

				store.EXPECT().
					ListEmployers(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(employers, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployers(t, recorder.Body, employers)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEmployers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEmployers(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.Employer{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEmployers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEmployers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store, nil, nil, nil, nil)
			recorder := httptest.NewRecorder()

			// Prepare request URL with query parameters
			url := "/employers"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomEmployer(username string) db.Employer {
	return db.Employer{
		ID:                  util.RandomInt(1, 1000),
		Username:            username,
		BusinessName:        util.RandomString(10),
		BusinessEmail:       util.RandomEmail(),
		BusinessPhone:       util.RandomPhoneNumber(),
		Location:            util.RandomString(20),
		Industry:            util.RandomString(10),
		ProfilePhoto:        util.RandomString(10),
		BusinessDescription: util.RandomString(50),
	}
}

func TestUpdateEmployersAPI(t *testing.T) {
	user, _ := randomUser(t)
	employer := randomEmployer(user.Username)
	updateReq := updateEmployerRequest{
		BusinessName:        "default_business_name",
		BusinessEmail:       "default_business_email",
		BusinessPhone:       "default_business_phone",
		Location:            "default_location",
		Industry:            "default_industry",
		ProfilePhoto:        "default_profile_photo",
		BusinessDescription: "default_business_description",
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"business_name":        updateReq.BusinessName,
				"business_email":       updateReq.BusinessEmail,
				"business_phone":       updateReq.BusinessPhone,
				"location":             updateReq.Location,
				"industry":             updateReq.Industry,
				"profile_photo":        updateReq.ProfilePhoto,
				"business_description": updateReq.BusinessDescription,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateEmployerParams{
					ID: employer.ID,
					BusinessName: pgtype.Text{
						String: updateReq.BusinessName,
						Valid:  updateReq.BusinessName != "",
					},
					BusinessEmail: pgtype.Text{
						String: updateReq.BusinessEmail,
						Valid:  updateReq.BusinessEmail != "",
					},
					BusinessPhone: pgtype.Text{
						String: updateReq.BusinessPhone,
						Valid:  updateReq.BusinessPhone != "",
					},
					Location: pgtype.Text{
						String: updateReq.Location,
						Valid:  updateReq.Location != "",
					},
					Industry: pgtype.Text{
						String: updateReq.Industry,
						Valid:  updateReq.Industry != "",
					},
					ProfilePhoto: pgtype.Text{
						String: updateReq.ProfilePhoto,
						Valid:  updateReq.ProfilePhoto != "",
					},
					BusinessDescription: pgtype.Text{
						String: updateReq.BusinessDescription,
						Valid:  updateReq.BusinessDescription != "",
					},
				}
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(employer, nil)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployer(t, recorder.Body, employer)
			},
		},
		{
			name: "NoAuthorization",
			body: gin.H{
				"BusinessName":        updateReq.BusinessName,
				"BusinessEmail":       updateReq.BusinessEmail,
				"BusinessPhone":       updateReq.BusinessPhone,
				"location":            updateReq.Location,
				"Industry":            updateReq.Industry,
				"profile_photo":       updateReq.ProfilePhoto,
				"BusinessDescription": updateReq.BusinessDescription,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateEmployer(gomock.Any(), gomock.Any()).
					Times(0)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			server := newTestServer(t, store, nil, nil, nil, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/employers/%d", employer.ID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchEmployer(t *testing.T, body *bytes.Buffer, employer db.Employer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEmployer db.Employer
	err = json.Unmarshal(data, &gotEmployer)
	require.NoError(t, err)

	// Ignore differences in the CreatedAt field's ext and wall fields
	require.Equal(t, employer.ID, gotEmployer.ID)
	require.Equal(t, employer.BusinessName, gotEmployer.BusinessName)
	require.Equal(t, employer.BusinessEmail, gotEmployer.BusinessEmail)
	require.Equal(t, employer.BusinessPhone, gotEmployer.BusinessPhone)
	require.Equal(t, employer.Location, gotEmployer.Location)
	require.Equal(t, employer.Industry, gotEmployer.Industry)
	require.Equal(t, employer.ProfilePhoto, gotEmployer.ProfilePhoto)
	require.Equal(t, employer.BusinessDescription, gotEmployer.BusinessDescription)
}

func requireBodyMatchEmployers(t *testing.T, body *bytes.Buffer, employers []db.Employer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEmployers []db.Employer
	err = json.Unmarshal(data, &gotEmployers)
	require.NoError(t, err)
	require.Equal(t, employers, gotEmployers)
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomString(6),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
		Role:           db.RoleEmployer,
	}
	return
}
