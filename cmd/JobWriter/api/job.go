package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
)

type createJobRequest struct {
	EmployerID     int64           `uri:"employer_id" binding:"required,min=1"`
	BusinessName   string          `json:"business_name"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Industry       string          `json:"industry"`
	JobLocation    string          `json:"job_location"`
	EmploymentType string          `json:"employment_type"`
	Wage           float32         `json:"wage"`
	Tips           float32         `json:"tips,omitempty"`
	JobApplication json.RawMessage `json:"job_application,omitempty"`
}

func createGooglePlaceIDQuery(jobLocation, businessName string) string {
	re := regexp.MustCompile(`,?\s*[A-Z]{2}\s*\d{5}(?:-\d{4})?$`)
	cleanedAddress := re.ReplaceAllString(jobLocation, "")
	cleanedAddress = strings.TrimSpace(cleanedAddress)
	if strings.HasSuffix(cleanedAddress, ",") {
		cleanedAddress = strings.TrimSuffix(cleanedAddress, ",")
	}
	return businessName + " " + cleanedAddress
}

func (server *Server) CreateJob(ctx *gin.Context) {
	var req createJobRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if authPayload.Role != db.RoleEmployer || authPayload.RoleID != req.EmployerID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account is not an employer")))
		return
	}

	placeID, err := server.gapi.GetPlaceID(createGooglePlaceIDQuery(req.JobLocation, req.BusinessName))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	data, err := server.gapi.GetPlaceDetails(placeID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := elasticsearch.Job{
		ID:                 fmt.Sprintf("%d_%s", authPayload.RoleID, req.Title),
		EmployerID:         authPayload.RoleID,
		HiringOrganization: req.BusinessName,
		Title:              req.Title,
		Industry:           req.Industry,
		JobLocation:        req.JobLocation,
		DatePosted:         time.Now().Format("2006-January-02"),
		Description:        req.Description,
		EmploymentType:     req.EmploymentType,
		Wage:               req.Wage,
		Tips:               req.Tips,
		JobApplication:     req.JobApplication,
		IsUserCreated:      true,
		PlaceID:            data.ID,
		DisplayName:        data.DisplayName.Text,
		PhoneNumber:        data.NationalPhoneNumber,
		BusinessType:       data.Types,
		FormattedAddress:   data.FormattedAddress,
		PreciseLocation: elasticsearch.GeoPoint{
			Lat: data.Location.Latitude,
			Lon: data.Location.Longitude,
		},
		Photos:        data.Photos,
		Rating:        data.Rating,
		PriceLevel:    data.PriceLevel,
		OpeningHours:  data.RegularOpeningHours,
		WebsiteURI:    data.WebsiteURI,
		GoogleMapsURI: data.GoogleMapsURI,
	}

	if err := server.esClient.IndexJob(&arg); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Job created successfully"})
}

type getJobRequest struct {
	JobID string `uri:"job_id" binding:"required"`
}

func (server *Server) GetJob(ctx *gin.Context) {
	var req getJobRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	job, err := server.esClient.GetJob(req.JobID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, job)
}

type updateJobRequest struct {
	JobID          string          `uri:"job_id" binding:"required,min=1"`
	EmployerID     int64           `json:"employer_id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Industry       string          `json:"industry"`
	JobLocation    string          `json:"job_location"`
	EmploymentType string          `json:"employment_type"`
	Wage           float32         `json:"wage"`
	Tips           float32         `json:"tips,omitempty"`
	JobApplication json.RawMessage `json:"job_application"`
}

func (server *Server) UpdateJob(ctx *gin.Context) {
	var req updateJobRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if authPayload.Role != db.RoleEmployer || authPayload.RoleID != req.EmployerID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account is not an employer")))
		return
	}

	arg := elasticsearch.Job{
		ID:             req.JobID,
		EmployerID:     req.EmployerID,
		Title:          req.Title,
		Description:    req.Description,
		JobLocation:    req.JobLocation,
		Industry:       req.Industry,
		EmploymentType: req.EmploymentType,
		Wage:           req.Wage,
		Tips:           req.Tips,
		JobApplication: req.JobApplication,
	}
	if err := server.esClient.UpdateJob(req.JobID, &arg); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Job updated successfully"})
}

type deleteJobRequest struct {
	JobID string `uri:"job_id" binding:"required"`
}

func (server *Server) DeleteJob(ctx *gin.Context) {
	var req deleteJobRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.esClient.DeleteJob(req.JobID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Job deleted successfully"})
}
