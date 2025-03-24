package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
)

type createCandidateRequest struct {
	FullName           string           `json:"full_name" binding:"required"`
	Email              string           `json:"email" binding:"required"`
	PhoneNumber        string           `json:"phone_number" binding:"required"`
	Education          es.Education     `json:"education"`
	Location           string           `json:"location"`
	SkillSet           []string         `json:"skill_sets"`
	Certificates       []string         `json:"certificates"`
	IndustryOfInterest string           `json:"industry_of_interest"`
	JobPreference      es.JobPreference `json:"job_preference"`
	TimeAvailability   []byte           `json:"time_availability"`
	ResumeFile         string           `json:"resume_file"`
	ProfilePhoto       string           `json:"profile_photo"`
	Description        string           `json:"description"`
}

func (server *Server) createCandidate(ctx *gin.Context) {
	var req createCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload, exists := ctx.MustGet(middleware.AuthorizationPayloadKey).(map[string]interface{})
	if !exists {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("authorization payload missing or invalid")))
		return
	}
	uid, ok := authPayload["uid"].(string)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("UID not found in authorization payload")))
		return
	}

	candidate := es.Candidate{
		UserUid:            uid,
		FullName:           req.FullName,
		Email:              req.Email,
		PhoneNumber:        req.PhoneNumber,
		Education:          req.Education,
		Location:           req.Location,
		SkillSet:           req.SkillSet,
		Certificates:       req.Certificates,
		IndustryOfInterest: req.IndustryOfInterest,
		JobPreference:      req.JobPreference,
		TimeAvailability:   req.TimeAvailability,
		ResumeFile:         req.ResumeFile,
		ProfilePhoto:       req.ProfilePhoto,
		Description:        req.Description,
		CreatedAt:          time.Time{},
	}

	if err := server.esClient.IndexCandidateV2(ctx, candidate); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, statusResponse("candidate indexed successfully"))
}

type getCandidateRequest struct {
	UID string `json:"uid" binding:"required,min=1"`
}

func (server *Server) getCandidate(ctx *gin.Context) {
	var req getCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	candidate, err := server.esClient.GetCandidate(ctx, req.UID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if authUID != req.UID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("not allowed to retrieve this resource: %v", req.UID)))
		return
	}
	ctx.JSON(http.StatusOK, candidate)
}

type listCandidatesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listCandidates(ctx *gin.Context) {
	var req listCandidatesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	arg := db.ListCandidatesParams{
		Username: authPayload.Username,
		Limit:    req.PageSize,
		Offset:   (req.PageID - 1) * req.PageSize,
	}

	candidates, err := server.store.ListCandidates(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, candidates)
}

type updateCandidateRequest struct {
	UserUID            string           `json:"uid" binding:"required,min=1"`
	FullName           string           `json:"full_name"`
	PhoneNumber        string           `json:"phone_number"`
	Education          db.Education     `json:"education"`
	Location           string           `json:"location"`
	SkillSet           []string         `json:"skill_set"`
	Certificate        []string         `json:"certificates"`
	IndustryOfInterest string           `json:"industry_of_interest"`
	JobPreference      db.JobPreference `json:"job_preference"`
	TimeAvailability   []byte           `json:"time_availability"`
	ResumeFile         string           `json:"resume_file"`
	ProfilePhoto       string           `json:"profile_photo"`
	Description        string           `json:"description"`
}

func (server *Server) updateCandidate(ctx *gin.Context) {
	var req updateCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if authUID != req.UserUID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("not allowed to update this resource: %v", req.UserUID)))
		return
	}
	data, err := json.Marshal(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	var updateFields map[string]interface{}
	err = json.Unmarshal(data, &updateFields)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if err := server.esClient.UpdateCandidateV2(ctx, req.UserUID, updateFields); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, statusResponse("candidate updated successfully"))
}

type deleteCandidateRequest struct {
	UID string `json:"uid" binding:"required,min=1"`
}

func (server *Server) deleteCandidate(ctx *gin.Context) {
	var req deleteCandidateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	if authUID != req.UID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(fmt.Errorf("not allowed to delete this resource: %v", req.UID)))
		return
	}

	if err := server.esClient.DeleteCandidate(ctx, req.UID); err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

}
