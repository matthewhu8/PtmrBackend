package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/hankimmy/PtmrBackend/pkg/db/mock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	mockes "github.com/hankimmy/PtmrBackend/pkg/elasticsearch/mock"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	mockwk "github.com/hankimmy/PtmrBackend/pkg/worker/mock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCreateCandidateApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleCandidate)
	appDoc := gin.H{"key1": "value1", "key2": 2}
	application := db.RandomCandidateApplication(1)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				"job_doc_id":         application.JobDocID,
				"application_doc":    appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, application.CandidateID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateCandidateApplicationTxParams{
					CreateCandidateApplicationParams: db.CreateCandidateApplicationParams{
						CandidateID:        application.CandidateID,
						EmployerID:         application.EmployerID,
						ElasticsearchDocID: application.ElasticsearchDocID,
						JobDocID:           application.JobDocID,
						ApplicationStatus:  application.ApplicationStatus,
					},
					AppDoc: appDoc,
				}
				store.EXPECT().
					CreateCandidateApplicationTx(gomock.Any(), EqCreateCandidateApplicationTxParams(arg)).
					Times(1).
					Return(db.CandidateAppTxResult{CandidateApplication: application}, nil)
				taskPayload := &worker.PayloadCreateApplication{
					DocID:  application.ElasticsearchDocID,
					AppDoc: appDoc,
				}
				taskDistributor.EXPECT().
					DistributeTaskCreateCandidateApplication(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateApplication(t, recorder.Body, application)
			},
		},
		{
			name: "MissingApplicationStatus",
			body: gin.H{
				"candidate_id":    application.CandidateID,
				"employer_id":     application.EmployerID,
				"application_doc": appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, application.CandidateID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUserRole",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				"application_doc":    appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, application.EmployerID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidJSON",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				// application_doc is set to a string instead of a JSON object
				"application_doc": "invalid_json",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, application.CandidateID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				"application_doc":    appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, application.CandidateID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateCandidateApplicationTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CandidateAppTxResult{}, sql.ErrConnDone)
				taskDistributor.EXPECT().
					DistributeTaskCreateCandidateApplication(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/candidate_applications"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateEmployerApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	application := db.RandomEmployerApplication(1)
	appDoc := gin.H{"message": application.Message}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				"application_doc":    appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, application.EmployerID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateEmployerApplicationParams{
					EmployerID:        application.EmployerID,
					CandidateID:       application.CandidateID,
					Message:           application.Message,
					ApplicationStatus: application.ApplicationStatus,
				}
				store.EXPECT().
					CreateEmployerApplication(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.EmployerApplication{
						EmployerID:        application.EmployerID,
						CandidateID:       application.CandidateID,
						Message:           application.Message,
						ApplicationStatus: application.ApplicationStatus}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployerApplication(t, recorder.Body, application)
			},
		},
		{
			name: "MissingApplicationStatus",
			body: gin.H{
				"candidate_id":    application.CandidateID,
				"employer_id":     application.EmployerID,
				"application_doc": appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, application.EmployerID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUserRole",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				"application_doc":    appDoc,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, application.CandidateID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidJSON",
			body: gin.H{
				"candidate_id":       application.CandidateID,
				"employer_id":        application.EmployerID,
				"application_status": application.ApplicationStatus,
				// application_doc is set to a string instead of a JSON object
				"application_doc": "invalid_json",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, application.EmployerID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/employer_applications"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetCandidateApplicationsByEmployerAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleCandidate)
	employer := db.RandomEmployer(user.Username)
	application := db.RandomCandidateApplication(employer.ID)
	docID := fmt.Sprintf("%d_%d", application.CandidateID, employer.ID)

	testCases := []struct {
		name          string
		employerID    int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, employer.ID)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				store.EXPECT().
					GetCandidateApplicationsByEmployer(gomock.Any(), gomock.Eq(employer.ID)).
					Times(1).
					Return([]db.CandidateApplication{application}, nil)
				var res map[string]interface{}
				esClient.EXPECT().
					GetCandidateApplication(gomock.Any(), gomock.Eq(docID)).
					Times(1).
					Return(res, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "NoAuthorization",
			employerID: employer.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				store.EXPECT().
					GetCandidateApplicationsByEmployer(gomock.Any(), gomock.Any()).
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
			mockEsClient := mockes.NewMockESClient(ctrl)
			tc.buildStubs(store, mockEsClient)
			server := newTestServer(t, store, mockEsClient, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/candidate_applications/%d", tc.employerID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetEmployerApplicationsByCandidateAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	candidate := db.RandomCandidate(user.Username)
	application := db.RandomEmployerApplication(candidate.ID)

	testCases := []struct {
		name          string
		candidateID   int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				store.EXPECT().
					GetEmployerApplicationsByCandidate(gomock.Any(), gomock.Eq(candidate.ID)).
					Times(1).
					Return([]db.EmployerApplication{application}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:        "NoAuthorization",
			candidateID: candidate.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				store.EXPECT().
					GetEmployerApplicationsByCandidate(gomock.Any(), gomock.Any()).
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
			mockEsClient := mockes.NewMockESClient(ctrl)
			tc.buildStubs(store, mockEsClient)

			server := newTestServer(t, store, mockEsClient, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/employer_applications/%d", tc.candidateID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateCandidateApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleCandidate)
	candidate := db.RandomCandidate(user.Username)
	docID := fmt.Sprintf("%d_%d", candidate.ID, 1)

	testCases := []struct {
		name          string
		candidateID   int64
		employerID    int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: candidate.ID,
			employerID:  1,
			body: gin.H{
				"application_status": db.ApplicationStatusRejected,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, 1)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				arg := db.UpdateCandidateApplicationParams{
					CandidateID: candidate.ID,
					EmployerID:  1,
					ApplicationStatus: db.NullApplicationStatus{
						ApplicationStatus: db.ApplicationStatusRejected,
						Valid:             true,
					},
					ElasticsearchDocID: pgtype.Text{
						String: docID,
						Valid:  true,
					},
				}

				store.EXPECT().
					UpdateCandidateApplication(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CandidateApplication{
						CandidateID:        candidate.ID,
						EmployerID:         1,
						ElasticsearchDocID: docID,
						ApplicationStatus:  db.ApplicationStatusRejected,
					}, nil)
				esClient.EXPECT().
					UpdateCandidateApplication(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateApplication(t, recorder.Body, db.CandidateApplication{
					CandidateID:        candidate.ID,
					EmployerID:         1,
					ElasticsearchDocID: docID,
					ApplicationStatus:  db.ApplicationStatusRejected,
				})
			},
		},
		{
			name:        "Update With Application",
			candidateID: candidate.ID,
			employerID:  1,
			body: gin.H{
				"application_status": db.ApplicationStatusSubmitted,
				"application_doc":    gin.H{"key1": "value1", "key2": 2},
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				arg := db.UpdateCandidateApplicationParams{
					CandidateID: candidate.ID,
					EmployerID:  1,
					ApplicationStatus: db.NullApplicationStatus{
						ApplicationStatus: db.ApplicationStatusSubmitted,
						Valid:             true,
					},
					ElasticsearchDocID: pgtype.Text{
						String: docID,
						Valid:  true,
					},
				}
				store.EXPECT().
					UpdateCandidateApplication(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.CandidateApplication{
						CandidateID:        candidate.ID,
						EmployerID:         1,
						ElasticsearchDocID: docID,
						ApplicationStatus:  db.ApplicationStatusSubmitted,
					}, nil)
				esClient.EXPECT().
					UpdateCandidateApplication(gomock.Any(), gomock.Eq(docID), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidateApplication(t, recorder.Body, db.CandidateApplication{
					CandidateID:        candidate.ID,
					EmployerID:         1,
					ElasticsearchDocID: docID,
					ApplicationStatus:  db.ApplicationStatusSubmitted,
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			mockEsClient := mockes.NewMockESClient(ctrl)
			tc.buildStubs(store, mockEsClient)

			server := newTestServer(t, store, mockEsClient, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/candidate_applications/%d/%d", tc.candidateID, tc.employerID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateEmployerApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	app := db.RandomEmployerApplication(1)

	testCases := []struct {
		name          string
		candidateID   int64
		employerID    int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: 1,
			employerID:  app.EmployerID,
			body: gin.H{
				"application_status": db.ApplicationStatusRejected,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, 1)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				arg := db.UpdateEmployerApplicationParams{
					CandidateID: 1,
					EmployerID:  app.EmployerID,
					ApplicationStatus: db.NullApplicationStatus{
						ApplicationStatus: db.ApplicationStatusRejected,
						Valid:             true,
					},
				}

				store.EXPECT().
					UpdateEmployerApplication(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.EmployerApplication{
						CandidateID:       1,
						EmployerID:        app.EmployerID,
						Message:           app.Message,
						ApplicationStatus: db.ApplicationStatusRejected,
					}, nil)
				esClient.EXPECT().
					UpdateEmployerApplication(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployerApplication(t, recorder.Body, db.EmployerApplication{
					CandidateID:       1,
					EmployerID:        app.EmployerID,
					Message:           app.Message,
					ApplicationStatus: db.ApplicationStatusRejected,
				})
			},
		},
		{
			name:        "Update With Application",
			employerID:  app.EmployerID,
			candidateID: 1,
			body: gin.H{
				"application_status": db.ApplicationStatusSubmitted,
				"application_doc":    gin.H{"key1": "value1", "key2": 2},
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, app.EmployerID)
			},
			buildStubs: func(store *mockdb.MockStore, esClient *mockes.MockESClient) {
				arg := db.UpdateEmployerApplicationParams{
					CandidateID: 1,
					EmployerID:  app.EmployerID,
					ApplicationStatus: db.NullApplicationStatus{
						ApplicationStatus: db.ApplicationStatusSubmitted,
						Valid:             true,
					},
				}
				store.EXPECT().
					UpdateEmployerApplication(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.EmployerApplication{
						CandidateID:       1,
						EmployerID:        app.EmployerID,
						Message:           app.Message,
						ApplicationStatus: db.ApplicationStatusSubmitted,
					}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEmployerApplication(t, recorder.Body, db.EmployerApplication{
					CandidateID:       1,
					EmployerID:        app.EmployerID,
					Message:           app.Message,
					ApplicationStatus: db.ApplicationStatusSubmitted,
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			mockEsClient := mockes.NewMockESClient(ctrl)
			tc.buildStubs(store, mockEsClient)

			server := newTestServer(t, store, mockEsClient, nil)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/employer_applications/%d/%d", tc.employerID, tc.candidateID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteCandidateApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleCandidate)
	candidate := db.RandomCandidate(user.Username)
	docID := fmt.Sprintf("%d_%d", candidate.ID, 1)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"candidate_id": candidate.ID,
				"employer_id":  1,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.DeleteApplicationTxParams{
					DeleteCandidateApplicationParams: db.DeleteCandidateApplicationParams{
						CandidateID: candidate.ID,
						EmployerID:  1,
					},
					IsEmployer: false,
					DocID:      docID,
				}
				store.EXPECT().
					DeleteApplicationTx(gomock.Any(), EqDeleteApplicationTxParams(arg)).
					Times(1).
					Return(nil)
				taskPayload := &worker.PayloadDeleteApplication{DocID: docID}
				taskDistributor.EXPECT().
					DistributeTaskDeleteCandidateApplication(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUserRole",
			body: gin.H{
				"candidate_id": candidate.ID,
				"employer_id":  1,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidJSONBody",
			body: gin.H{
				"candidate_id": candidate.ID,
				"employer_id":  "invalid_id", // Passing a string instead of an int64
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidIDs",
			body: gin.H{
				"candidate_id": 0,  // Invalid ID
				"employer_id":  -1, // Invalid ID
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: gin.H{
				"candidate_id": candidate.ID,
				"employer_id":  1,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					DeleteApplicationTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
				taskDistributor.EXPECT().
					DistributeTaskDeleteCandidateApplication(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()
			data, _ := json.Marshal(tc.body)
			url := fmt.Sprintf("/candidate_applications/%s", candidate.Username)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteEmployerApplicationAPI(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	employer := db.RandomEmployer(user.Username)
	docID := fmt.Sprintf("%d_%d", employer.ID, 1)

	testCases := []struct {
		name             string
		employerUsername string
		body             gin.H
		setupAuth        func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs       func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse    func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:             "OK",
			employerUsername: user.Username,
			body: gin.H{
				"candidate_id": 1,
				"employer_id":  employer.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.DeleteApplicationTxParams{
					DeleteEmployerApplicationParams: db.DeleteEmployerApplicationParams{
						CandidateID: 1,
						EmployerID:  employer.ID,
					},
					IsEmployer: true,
					DocID:      docID,
				}
				store.EXPECT().
					DeleteApplicationTx(gomock.Any(), EqDeleteApplicationTxParams(arg)).
					Times(1).
					Return(nil)
				taskPayload := &worker.PayloadDeleteApplication{DocID: docID}
				taskDistributor.EXPECT().
					DistributeTaskDeleteEmployerApplication(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(store, taskDistributor)

			server := newTestServer(t, store, nil, taskDistributor)
			recorder := httptest.NewRecorder()
			data, _ := json.Marshal(tc.body)
			url := fmt.Sprintf("/employer_applications/%s", tc.employerUsername)
			request, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchCandidateApplication(t *testing.T, body *bytes.Buffer, application db.CandidateApplication) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotApplication db.CandidateApplication
	err = json.Unmarshal(data, &gotApplication)
	require.NoError(t, err)

	require.Equal(t, application.CandidateID, gotApplication.CandidateID)
	require.Equal(t, application.EmployerID, gotApplication.EmployerID)
	require.Equal(t, application.ElasticsearchDocID, gotApplication.ElasticsearchDocID)
	require.Equal(t, application.ApplicationStatus, gotApplication.ApplicationStatus)
}

func requireBodyMatchEmployerApplication(t *testing.T, body *bytes.Buffer, application db.EmployerApplication) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotApplication db.EmployerApplication
	err = json.Unmarshal(data, &gotApplication)
	require.NoError(t, err)

	require.Equal(t, application.CandidateID, gotApplication.CandidateID)
	require.Equal(t, application.EmployerID, gotApplication.EmployerID)
	require.Equal(t, application.Message, gotApplication.Message)
	require.Equal(t, application.ApplicationStatus, gotApplication.ApplicationStatus)
}

type eqCreateCandidateApplicationTxParamsMatcher struct {
	arg db.CreateCandidateApplicationTxParams
}

func (expected eqCreateCandidateApplicationTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateCandidateApplicationTxParams)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(expected.arg.CreateCandidateApplicationParams, actualArg.CreateCandidateApplicationParams) {
		return false
	}
	err := actualArg.AfterCreate(expected.arg.CreateCandidateApplicationParams.ElasticsearchDocID, expected.arg.AppDoc)
	return err == nil
}

func (expected eqCreateCandidateApplicationTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", expected.arg)
}

func EqCreateCandidateApplicationTxParams(arg db.CreateCandidateApplicationTxParams) gomock.Matcher {
	return eqCreateCandidateApplicationTxParamsMatcher{arg}
}

type eqDeleteApplicationTxParamsMatcher struct {
	arg db.DeleteApplicationTxParams
}

func (expected eqDeleteApplicationTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.DeleteApplicationTxParams)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(expected.arg.DeleteEmployerApplicationParams, actualArg.DeleteEmployerApplicationParams) &&
		!reflect.DeepEqual(expected.arg.DeleteCandidateApplicationParams, actualArg.DeleteCandidateApplicationParams) {
		return false
	}
	err := actualArg.AfterDelete(expected.arg.DocID)
	return err == nil
}

func (expected eqDeleteApplicationTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", expected.arg)
}

func EqDeleteApplicationTxParams(arg db.DeleteApplicationTxParams) gomock.Matcher {
	return eqDeleteApplicationTxParamsMatcher{arg}
}
