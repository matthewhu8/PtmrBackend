package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	mockes "github.com/hankimmy/PtmrBackend/pkg/elasticsearch/mock"
	mockgapi "github.com/hankimmy/PtmrBackend/pkg/google/mock"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestGetCandidateBatchFeed(t *testing.T) {
	user, _ := db.RandomUser(db.RoleCandidate)
	candidate := db.RandomCandidate(user.Username)
	candidateBody := gin.H{
		"industry":        "Tech",
		"employment_type": "Full-time",
		"title":           "Software Engineer",
		"location":        "13 E 37th St, New York, NY",
		"distance":        "10mi",
	}
	expectedJobs := []elasticsearch.Job{
		elasticsearch.RandomJob(candidate.ID),
		elasticsearch.RandomJob(candidate.ID),
	}

	testCases := []struct {
		name          string
		candidateID   int64
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:        "OK",
			candidateID: candidate.ID,
			body:        candidateBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, candidate.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
				gapi.EXPECT().
					GetLatLon(gomock.Eq(candidateBody["location"].(string))).
					Times(1).
					Return(40.7501259, -73.9820676, nil)
				esClient.EXPECT().
					SearchJobs(gomock.Eq(candidateBody["industry"]),
						gomock.Eq(candidateBody["employment_type"]),
						gomock.Eq(candidateBody["title"]),
						gomock.Eq(candidateBody["distance"]),
						gomock.Eq(elasticsearch.GeoPoint{
							Lat: 40.7501259,
							Lon: -73.9820676,
						})).
					Times(1).
					Return(expectedJobs, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchJobs(t, recorder.Body, expectedJobs)
			},
		},
		{
			name:        "UnauthorizedUserRole",
			candidateID: candidate.ID,
			body:        candidateBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, candidate.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:        "InvalidJSONBody",
			candidateID: candidate.ID,
			body: gin.H{
				"industry": 123, // Invalid data type for industry
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, candidate.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:        "InternalServerError",
			candidateID: candidate.ID,
			body:        candidateBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, candidate.Username, db.RoleCandidate, time.Minute, candidate.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
				gapi.EXPECT().
					GetLatLon(gomock.Eq(candidateBody["location"].(string))).
					Times(1).
					Return(40.7501259, -73.9820676, nil)
				esClient.EXPECT().
					SearchJobs(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Times(1).
					Return(nil, errors.New("internal server error"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			esCtrl := gomock.NewController(t)
			defer esCtrl.Finish()
			esClient := mockes.NewMockESClient(esCtrl)
			gCtrl := gomock.NewController(t)
			defer gCtrl.Finish()
			gClient := mockgapi.NewMockGAPI(gCtrl)
			tc.buildStubs(esClient, gClient)

			server := newTestServer(t, esClient, gClient)
			recorder := httptest.NewRecorder()
			data, _ := json.Marshal(tc.body)
			url := fmt.Sprintf("/feed/%d", tc.candidateID)
			request, err := http.NewRequest(http.MethodGet, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchJobs(t *testing.T, body *bytes.Buffer, expectedJobs []elasticsearch.Job) {
	var gotJobs []elasticsearch.Job
	err := json.Unmarshal(body.Bytes(), &gotJobs)
	require.NoError(t, err)

	require.Equal(t, len(expectedJobs), len(gotJobs))

	for i, job := range expectedJobs {
		require.Equal(t, job.ID, gotJobs[i].ID)
		require.Equal(t, job.EmployerID, gotJobs[i].EmployerID)
		require.Equal(t, job.Title, gotJobs[i].Title)
		require.Equal(t, job.Description, gotJobs[i].Description)
		require.Equal(t, job.JobLocation, gotJobs[i].JobLocation)
		require.Equal(t, job.Industry, gotJobs[i].Industry)
		require.Equal(t, job.EmploymentType, gotJobs[i].EmploymentType)
		require.Equal(t, job.Wage, gotJobs[i].Wage)
		require.Equal(t, job.Tips, gotJobs[i].Tips)
		require.Equal(t, job.DatePosted, gotJobs[i].DatePosted)
		require.Equal(t, job.IsUserCreated, gotJobs[i].IsUserCreated)
	}
}
