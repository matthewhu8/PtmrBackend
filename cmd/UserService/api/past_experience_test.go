package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	mockes "github.com/hankimmy/PtmrBackend/pkg/elasticsearch/mock"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	mockwk "github.com/hankimmy/PtmrBackend/pkg/worker/mock"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestGetPastExperienceAPI(t *testing.T) {
	candidate := randomCandidate()
	pastExperience := randomPastExperience()
	esCtrl := gomock.NewController(t)
	defer esCtrl.Finish()
	esClient := mockes.NewMockESClient(esCtrl)

	auth := mockCandidateMiddleware(t, candidate.UserUid)
	type Query struct {
		userUID string
		id      string
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				userUID: candidate.UserUid,
				id:      pastExperience.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					GetPastExperience(gomock.Any(), gomock.Eq(candidate.UserUid), gomock.Eq(pastExperience.ID)).
					Times(1).
					Return(&pastExperience, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPastExperience(t, recorder.Body, pastExperience)
			},
		},
		{
			name: "NoAuthorization",
			query: Query{
				userUID: candidate.UserUid,
				id:      pastExperience.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(esClient *mockes.MockESClient) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "Invalid ID",
			query: Query{
				userUID: candidate.UserUid,
				id:      "**@123",
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Internal Server Error",
			query: Query{
				userUID: candidate.UserUid,
				id:      pastExperience.ID,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					GetPastExperience(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, errors.New("failed to get document"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(esClient)

			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/candidates/past_experience/")
			request, _ := http.NewRequest(http.MethodGet, url, nil)
			q := request.URL.Query()
			q.Add("user_uid", tc.query.userUID)
			q.Add("id", tc.query.id)
			request.URL.RawQuery = q.Encode()
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreatePastExperienceAPI(t *testing.T) {
	candidate := randomCandidate()
	pastExperience := randomPastExperience()
	auth := mockCandidateMiddleware(t, candidate.UserUid)

	reqBody := gin.H{
		"user_uid":    candidate.UserUid,
		"industry":    pastExperience.Industry,
		"employer":    pastExperience.Employer,
		"job_title":   pastExperience.JobTitle,
		"start_date":  pastExperience.StartDate,
		"end_date":    pastExperience.EndDate,
		"present":     pastExperience.Present,
		"description": pastExperience.Description,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskAddPastExperience(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, payload *worker.PayloadPastExperience, opts ...asynq.Option) error {
						require.Equal(t, candidate.UserUid, payload.UserUID)
						require.Equal(t, pastExperience.Industry, payload.PastExperience.Industry)
						require.Equal(t, pastExperience.Employer, payload.PastExperience.Employer)
						require.Equal(t, pastExperience.JobTitle, payload.PastExperience.JobTitle)
						require.Equal(t, pastExperience.StartDate, payload.PastExperience.StartDate)
						require.Equal(t, pastExperience.EndDate, payload.PastExperience.EndDate)
						require.Equal(t, pastExperience.Present, payload.PastExperience.Present)
						require.Equal(t, pastExperience.Description, payload.PastExperience.Description)
						return nil
					}).Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchStatus(t, recorder.Body, "task queued successfully")
			},
		},
		{
			name: "NoAuthorization",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				// No authorization is provided
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				// No expectations for the taskDistributor since the request is unauthorized
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				// Expect the response code to be 401 Unauthorized
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(taskDistributor)

			server := newTestServer(t, nil, taskDistributor, nil, auth, nil)
			recorder := httptest.NewRecorder()
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/candidates/past_experience/"
			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestListPastExperiencesAPI(t *testing.T) {
	candidate := randomCandidate()

	auth := mockCandidateMiddleware(t, candidate.UserUid)
	n := 5
	pastExperiences := make([]es.PastExperience, n)
	for i := 0; i < n; i++ {
		pastExperiences[i] = randomPastExperience()
	}

	type Query struct {
		userUID string
	}

	testQuery := Query{
		userUID: candidate.UserUid,
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: testQuery,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().ListPastExperiences(gomock.Any(), gomock.Eq(candidate.UserUid)).
					Times(1).
					Return(pastExperiences, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchPastExperiences(t, recorder.Body, pastExperiences)
			},
		},
		{
			name:  "NoAuthorization",
			query: testQuery,
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(esClient *mockes.MockESClient) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:  "InternalError",
			query: testQuery,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					ListPastExperiences(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, errors.New("fail to return all past experiences"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:  "Invalid Query",
			query: Query{userUID: "234&48*@"},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					ListPastExperiences(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			esCtrl := gomock.NewController(t)
			defer esCtrl.Finish()
			esClient := mockes.NewMockESClient(esCtrl)
			tc.buildStubs(esClient)
			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()

			url := "/candidates/past_experience/all"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("user_uid", tc.query.userUID)
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdatePastExperienceAPI(t *testing.T) {
	candidate := randomCandidate()
	pastExperience := randomPastExperience()
	auth := mockCandidateMiddleware(t, candidate.UserUid)

	newIndustry := util.RandomString(10)
	newJobTitle := util.RandomString(10)

	reqBody := gin.H{
		"user_uid":  candidate.UserUid,
		"id":        pastExperience.ID,
		"industry":  newIndustry,
		"job_title": newJobTitle,
	}

	updateFields := es.PastExperience{
		ID:          pastExperience.ID,
		Industry:    newIndustry,
		Employer:    "",
		JobTitle:    newJobTitle,
		StartDate:   time.Time{},
		EndDate:     time.Time{},
		Present:     false,
		Description: "",
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskPayload := &worker.PayloadPastExperience{
					PastExperience: updateFields,
					UserUID:        candidate.UserUid,
				}
				taskDistributor.EXPECT().
					DistributeTaskUpdatePastExperience(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchStatus(t, recorder.Body, "update past experience task queued successfully")
			},
		},
		{
			name: "UnauthorizedUser",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskPayload := &worker.PayloadPastExperience{
					PastExperience: updateFields,
					UserUID:        candidate.UserUid,
				}
				taskDistributor.EXPECT().
					DistributeTaskUpdatePastExperience(gomock.Any(), taskPayload, gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskUpdatePastExperience(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("failed to distribute"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(taskDistributor)

			server := newTestServer(t, nil, taskDistributor, nil, auth, nil)
			recorder := httptest.NewRecorder()

			url := "/candidates/past_experience/"
			data := marshalRequestBody(t, tc.body)
			request := createNewRequest(t, http.MethodPatch, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeletePastExperienceAPI(t *testing.T) {
	candidate := randomCandidate()
	pastExperience := randomPastExperience()
	auth := mockCandidateMiddleware(t, candidate.UserUid)
	reqBody := gin.H{
		"user_uid": candidate.UserUid,
		"id":       pastExperience.ID,
	}
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskPayload := &worker.PayloadDeletePastExperience{
					PastExperienceID: pastExperience.ID,
					UserUID:          candidate.UserUid,
				}
				taskDistributor.EXPECT().
					DistributeTaskDeletePastExperience(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchStatus(t, recorder.Body, "delete past experience task queued successfully")
			},
		},
		{
			name: "UnauthorizedUser",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDeletePastExperience(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: reqBody,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(taskDistributor *mockwk.MockTaskDistributor) {
				taskDistributor.EXPECT().
					DistributeTaskDeletePastExperience(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).Return(errors.New("failed to distribute task delete"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			tc.buildStubs(taskDistributor)

			server := newTestServer(t, nil, taskDistributor, nil, auth, nil)
			recorder := httptest.NewRecorder()

			url := "/candidates/past_experience/"
			data := marshalRequestBody(t, tc.body)

			request := createNewRequest(t, http.MethodDelete, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomPastExperience() es.PastExperience {
	start, end := randomDateRange(2024)
	return es.PastExperience{
		ID:          util.RandomString(15),
		Industry:    util.RandomString(10),
		Employer:    util.RandomString(10),
		JobTitle:    util.RandomString(10),
		StartDate:   start,
		EndDate:     end,
		Present:     util.RandomBool(),
		Description: util.RandomString(20),
	}
}

func requireBodyMatchPastExperience(t *testing.T, body *bytes.Buffer, pastExperience es.PastExperience) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPastExperience es.PastExperience
	err = json.Unmarshal(data, &gotPastExperience)
	require.NoError(t, err)

	require.Equal(t, pastExperience.ID, gotPastExperience.ID)
	require.Equal(t, pastExperience.Industry, gotPastExperience.Industry)
	require.Equal(t, pastExperience.JobTitle, gotPastExperience.JobTitle)
	require.Equal(t, pastExperience.Employer, gotPastExperience.Employer)
	require.Equal(t, pastExperience.EndDate, gotPastExperience.EndDate)
	require.Equal(t, pastExperience.StartDate, gotPastExperience.StartDate)
	require.Equal(t, pastExperience.Present, gotPastExperience.Present)
	require.Equal(t, pastExperience.Description, gotPastExperience.Description)
}

func requireBodyMatchPastExperiences(t *testing.T, body *bytes.Buffer, pastExperiences []es.PastExperience) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotPastExperiences []es.PastExperience
	err = json.Unmarshal(data, &gotPastExperiences)
	require.NoError(t, err)

	sort.SliceStable(pastExperiences, func(i, j int) bool {
		return pastExperiences[i].StartDate.Before(pastExperiences[j].StartDate)
	})

	sort.SliceStable(gotPastExperiences, func(i, j int) bool {
		return gotPastExperiences[i].StartDate.Before(gotPastExperiences[j].StartDate)
	})

	require.Equal(t, pastExperiences, gotPastExperiences)
}

func randomDateRange(year int) (time.Time, time.Time) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	randomStartDate := randomDate(startOfYear, endOfYear).UTC().Truncate(24 * time.Hour)
	randomEndDate := randomDate(randomStartDate, endOfYear).UTC().Truncate(24 * time.Hour)

	return randomStartDate, randomEndDate
}

func randomDate(start, end time.Time) time.Time {
	delta := end.Unix() - start.Unix()
	sec := rand.Int63n(delta) + start.Unix()
	return time.Unix(sec, 0)
}

func requireBodyMatchStatus(t *testing.T, body *bytes.Buffer, expectedStatus string) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var got map[string]string
	err = json.Unmarshal(data, &got)
	require.NoError(t, err)

	require.Equal(t, expectedStatus, got["status"])
}
