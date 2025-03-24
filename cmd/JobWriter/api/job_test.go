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
	"github.com/hankimmy/PtmrBackend/pkg/google"
	mockgapi "github.com/hankimmy/PtmrBackend/pkg/google/mock"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestCreateJob(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	employer := db.RandomEmployer(user.Username)
	job := elasticsearch.RandomJob(employer.ID)
	jobBody := gin.H{
		"business_name":   job.HiringOrganization,
		"title":           job.Title,
		"description":     job.Description,
		"industry":        job.Industry,
		"job_location":    job.JobLocation,
		"employment_type": job.EmploymentType,
		"wage":            job.Wage,
		"tips":            job.Tips,
		"job_application": job.JobApplication,
	}
	arg := elasticsearch.Job{
		ID:                 job.ID,
		EmployerID:         employer.ID,
		HiringOrganization: job.HiringOrganization,
		Title:              job.Title,
		Description:        job.Description,
		JobLocation:        job.JobLocation,
		Industry:           job.Industry,
		EmploymentType:     job.EmploymentType,
		Wage:               job.Wage,
		Tips:               job.Tips,
		JobApplication:     job.JobApplication,
		DatePosted:         job.DatePosted,
		IsUserCreated:      job.IsUserCreated,
		PlaceID:            job.PlaceID,
		DisplayName:        job.DisplayName,
		PhoneNumber:        job.PhoneNumber,
		BusinessType:       job.BusinessType,
		FormattedAddress:   job.FormattedAddress,
		PreciseLocation:    job.PreciseLocation,
		Photos:             job.Photos,
		Rating:             job.Rating,
		PriceLevel:         job.PriceLevel,
		OpeningHours:       job.OpeningHours,
		WebsiteURI:         job.WebsiteURI,
		GoogleMapsURI:      job.GoogleMapsURI,
	}
	placeDetailsResponse := google.PlaceDetailsResponse{
		Name:             "",
		ID:               job.PlaceID,
		Types:            job.BusinessType,
		FormattedAddress: job.FormattedAddress,
		Location: google.Location{
			Latitude:  job.PreciseLocation.Lat,
			Longitude: job.PreciseLocation.Lon,
		},
		Photos:        job.Photos,
		GoogleMapsURI: job.GoogleMapsURI,
		DisplayName: struct {
			Text string `json:"text"`
		}(struct{ Text string }{Text: job.DisplayName}),
		NationalPhoneNumber: job.PhoneNumber,
		PriceLevel:          job.PriceLevel,
		Rating:              job.Rating,
		RegularOpeningHours: job.OpeningHours,
		WebsiteURI:          job.WebsiteURI,
	}
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: jobBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, employer.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
				gapi.EXPECT().
					GetPlaceID(gomock.Any()).
					Times(1).
					Return("ChIJJS3mqONZwokR9KlP3H_7MNg", nil)
				gapi.EXPECT().
					GetPlaceDetails(gomock.Any()).
					Times(1).
					Return(&placeDetailsResponse, nil)
				esClient.EXPECT().
					IndexJob(gomock.Eq(&arg)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUserRole",
			body: jobBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleCandidate, time.Minute, employer.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidJSONBody",
			body: gin.H{
				"title": 123, // Invalid data type for title
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, employer.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			body: jobBody,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, employer.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient, gapi *mockgapi.MockGAPI) {
				gapi.EXPECT().
					GetPlaceID(gomock.Any()).
					Times(1).
					Return("ChIJJS3mqONZwokR9KlP3H_7MNg", nil)
				gapi.EXPECT().
					GetPlaceDetails(gomock.Any()).
					Times(1).
					Return(&placeDetailsResponse, nil)
				esClient.EXPECT().
					IndexJob(gomock.Any()).
					Times(1).
					Return(errors.New(elasticsearch.ErrIndexFailure))
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
			url := fmt.Sprintf("/jobs/%d", employer.ID)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestGetJob(t *testing.T) {
	jobID := "job_123"
	job := elasticsearch.RandomJob(0)

	testCases := []struct {
		name          string
		jobID         string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: jobID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "", db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					GetJob(gomock.Eq(jobID)).
					Times(1).
					Return(&job, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.JSONEq(t, toJson(job), recorder.Body.String())
			},
		},
		{
			name:  "NotFound",
			jobID: "non_existent_job",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "", db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					GetJob(gomock.Eq("non_existent_job")).
					Times(1).
					Return(nil, errors.New(elasticsearch.ErrGetFailure))
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			esCtrl := gomock.NewController(t)
			defer esCtrl.Finish()
			esClient := mockes.NewMockESClient(esCtrl)
			tc.buildStubs(esClient)

			server := newTestServer(t, esClient, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/jobs/%s", tc.jobID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestUpdateJob(t *testing.T) {
	user, _ := db.RandomUser(db.RoleEmployer)
	employer := db.RandomEmployer(user.Username)
	jobID := "job_123"
	updatedJob := elasticsearch.RandomJob(0)

	testCases := []struct {
		name          string
		jobID         string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: jobID,
			body: gin.H{
				"employer_id":     employer.ID,
				"title":           updatedJob.Title,
				"description":     updatedJob.Description,
				"industry":        updatedJob.Industry,
				"employment_type": updatedJob.EmploymentType,
				"wage":            updatedJob.Wage,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, user.Username, db.RoleEmployer, time.Minute, employer.ID)
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					UpdateJob(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Contains(t, recorder.Body.String(), "Job updated successfully")
			},
		},
		{
			name:  "Unauthorized",
			jobID: jobID,
			body: gin.H{
				"employer_id":     employer.ID,
				"title":           updatedJob.Title,
				"description":     updatedJob.Description,
				"industry":        updatedJob.Industry,
				"job_location":    updatedJob.JobLocation,
				"employment_type": updatedJob.EmploymentType,
				"wage":            updatedJob.Wage,
				"tips":            updatedJob.Tips,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "", db.RoleCandidate, time.Minute, 0)
			},
			buildStubs: func(esClient *mockes.MockESClient) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			esCtrl := gomock.NewController(t)
			defer esCtrl.Finish()
			esClient := mockes.NewMockESClient(esCtrl)
			tc.buildStubs(esClient)
			server := newTestServer(t, esClient, nil)
			recorder := httptest.NewRecorder()

			data, _ := json.Marshal(tc.body)
			url := fmt.Sprintf("/jobs/%s", tc.jobID)
			request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteJob(t *testing.T) {
	jobID := "job_123"

	testCases := []struct {
		name          string
		jobID         string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(esClient *mockes.MockESClient)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			jobID: jobID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "", db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					DeleteJob(gomock.Eq(jobID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				require.Contains(t, recorder.Body.String(), "Job deleted successfully")
			},
		},
		{
			name:  "InternalServerError",
			jobID: jobID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				middleware.AddAuthorization(t, request, tokenMaker, middleware.AuthorizationTypeBearer, "", db.RoleEmployer, time.Minute, 0)
			},
			buildStubs: func(esClient *mockes.MockESClient) {
				esClient.EXPECT().
					DeleteJob(gomock.Eq(jobID)).
					Times(1).
					Return(errors.New("internal server error"))
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
			tc.buildStubs(esClient)
			server := newTestServer(t, esClient, nil)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/jobs/%s", tc.jobID)
			request, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func toJson(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "" // In real tests, you might want to handle errors differently
	}
	return string(bytes)
}
