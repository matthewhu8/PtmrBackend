package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	mockes "github.com/hankimmy/PtmrBackend/pkg/elasticsearch/mock"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestCreateCandidateAPI(t *testing.T) {
	candidate := randomCandidate()
	esCtrl := gomock.NewController(t)
	defer esCtrl.Finish()
	esClient := mockes.NewMockESClient(esCtrl)

	auth := mockCandidateMiddleware(t, candidate.UserUid)

	req := gin.H{
		"full_name":            candidate.FullName,
		"email":                candidate.Email,
		"phone_number":         candidate.PhoneNumber,
		"education":            candidate.Education,
		"location":             candidate.Location,
		"skill_sets":           candidate.SkillSet,
		"certificates":         candidate.Certificates,
		"industry_of_interest": candidate.IndustryOfInterest,
		"job_preference":       candidate.JobPreference,
		"time_availability":    candidate.TimeAvailability,
		"resume_file":          candidate.ResumeFile,
		"profile_photo":        candidate.ProfilePhoto,
		"description":          candidate.Description,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: req,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				arg := es.Candidate{
					UserUid:            candidate.UserUid,
					FullName:           candidate.FullName,
					Email:              candidate.Email,
					PhoneNumber:        candidate.PhoneNumber,
					Education:          candidate.Education,
					Location:           candidate.Location,
					SkillSet:           candidate.SkillSet,
					Certificates:       candidate.Certificates,
					IndustryOfInterest: candidate.IndustryOfInterest,
					JobPreference:      candidate.JobPreference,
					TimeAvailability:   candidate.TimeAvailability,
					ResumeFile:         candidate.ResumeFile,
					ProfilePhoto:       candidate.ProfilePhoto,
					Description:        candidate.Description,
				}
				esClient.EXPECT().
					IndexCandidateV2(gomock.Any(), arg).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			body: req,
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					IndexCandidateV2(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: req,
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					IndexCandidateV2(gomock.Any(), gomock.Any()).
					Times(1).
					Return(errors.New("index failed"))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(esClient)
			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)

			url := "/candidates/"
			request := createNewRequest(t, http.MethodPut, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateCandidateAPI(t *testing.T) {
	candidate := randomCandidate()
	esCtrl := gomock.NewController(t)
	defer esCtrl.Finish()
	esClient := mockes.NewMockESClient(esCtrl)

	auth := mockCandidateMiddleware(t, candidate.UserUid)

	var req updateCandidateRequest
	industry := "retail"
	req.UserUID = candidate.UserUid
	req.IndustryOfInterest = industry
	var updateFields map[string]interface{}
	data, _ := json.Marshal(req)
	json.Unmarshal(data, &updateFields)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"uid":                  candidate.UserUid,
				"industry_of_interest": industry,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					UpdateCandidateV2(gomock.Any(), candidate.UserUid, updateFields).
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
			tc.buildStubs(esClient)
			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)

			url := "/candidates/"
			request := createNewRequest(t, http.MethodPatch, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetCandidateAPI(t *testing.T) {
	candidate := randomCandidate()
	esCtrl := gomock.NewController(t)
	defer esCtrl.Finish()
	esClient := mockes.NewMockESClient(esCtrl)

	auth := mockCandidateMiddleware(t, candidate.UserUid)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"uid": candidate.UserUid,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					GetCandidate(gomock.Any(), candidate.UserUid).
					Times(1).
					Return(&candidate, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchCandidate(t, recorder.Body, candidate)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(esClient)
			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)

			url := "/candidates/"
			request := createNewRequest(t, http.MethodGet, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteCandidate(t *testing.T) {
	candidate := randomCandidate()
	esCtrl := gomock.NewController(t)
	defer esCtrl.Finish()
	esClient := mockes.NewMockESClient(esCtrl)

	auth := mockCandidateMiddleware(t, candidate.UserUid)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"uid": candidate.UserUid,
			},
			setupAuth: func(t *testing.T, request *http.Request) {
				firebase.AddAuthorization(t, request, firebase.AuthorizationTypeBearer, string(db.RoleCandidate))
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					DeleteCandidate(gomock.Any(), candidate.UserUid).
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
			tc.buildStubs(esClient)
			server := newTestServer(t, nil, nil, esClient, auth, nil)
			recorder := httptest.NewRecorder()

			data := marshalRequestBody(t, tc.body)

			url := "/candidates/"
			request := createNewRequest(t, http.MethodDelete, url, data)
			tc.setupAuth(t, request)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomCandidate() es.Candidate {
	return es.Candidate{
		UserUid:            util.RandomString(10),
		FullName:           util.RandomString(5),
		Email:              util.RandomEmail(),
		PhoneNumber:        util.RandomPhoneNumber(),
		Education:          es.EducationBachelor,
		Location:           util.RandomString(20),
		SkillSet:           []string{util.RandomString(5), util.RandomString(5)},
		Certificates:       []string{util.RandomString(5)},
		IndustryOfInterest: util.RandomString(10),
		JobPreference:      es.JobPreferenceInPerson,
		TimeAvailability:   util.RandomAvailability(),
		ResumeFile:         util.RandomString(10),
		ProfilePhoto:       util.RandomString(10),
		Description:        util.RandomString(50),
	}
}

func requireJSONEqual(t *testing.T, expected, actual []byte) {
	var expectedNormalized, actualNormalized interface{}

	err := json.Unmarshal(expected, &expectedNormalized)
	require.NoError(t, err)

	err = json.Unmarshal(actual, &actualNormalized)
	require.NoError(t, err)

	require.Equal(t, expectedNormalized, actualNormalized)
}

func requireBodyMatchCandidate(t *testing.T, body *bytes.Buffer, candidate es.Candidate) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotCandidate es.Candidate
	err = json.Unmarshal(data, &gotCandidate)
	require.NoError(t, err)

	require.Equal(t, candidate.UserUid, gotCandidate.UserUid)
	require.Equal(t, candidate.FullName, gotCandidate.FullName)
	require.Equal(t, candidate.PhoneNumber, gotCandidate.PhoneNumber)
	require.Equal(t, candidate.Education, gotCandidate.Education)
	require.Equal(t, candidate.Location, gotCandidate.Location)
	require.Equal(t, candidate.SkillSet, gotCandidate.SkillSet)
	require.Equal(t, candidate.Certificates, gotCandidate.Certificates)
	require.Equal(t, candidate.IndustryOfInterest, gotCandidate.IndustryOfInterest)
	require.Equal(t, candidate.JobPreference, gotCandidate.JobPreference)
	requireJSONEqual(t, candidate.TimeAvailability, gotCandidate.TimeAvailability)
	require.Equal(t, candidate.ResumeFile, gotCandidate.ResumeFile)
	require.Equal(t, candidate.ProfilePhoto, gotCandidate.ProfilePhoto)
	require.Equal(t, candidate.Description, gotCandidate.Description)
}
