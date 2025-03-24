package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	es "github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/hibiken/asynq"
)

type createPastExperienceRequest struct {
	UserUID     string    `json:"user_uid" binding:"required"`
	Industry    string    `json:"industry" binding:"required"`
	Employer    string    `json:"employer" binding:"required"`
	JobTitle    string    `json:"job_title" binding:"required"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Present     bool      `json:"present"`
	Description string    `json:"description" binding:"required"`
}

func (server *Server) createPastExperience(ctx *gin.Context) {
	var req createPastExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if authUID != req.UserUID || err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}
	data := []byte(req.UserUID + time.Now().String())
	hash := sha256.Sum256(data)
	pastExperience := es.PastExperience{
		ID:          hex.EncodeToString(hash[:]),
		Industry:    req.Industry,
		Employer:    req.Employer,
		JobTitle:    req.JobTitle,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Present:     req.Present,
		Description: req.Description,
	}
	payload := &worker.PayloadPastExperience{
		PastExperience: pastExperience,
		UserUID:        req.UserUID,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.ProcessIn(5 * time.Second),
		asynq.Queue(worker.QueueDefault),
	}
	err = server.taskDistributor.DistributeTaskAddPastExperience(ctx, payload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to distribute task: %w", err)))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "task queued successfully"})
}

type getPastExperienceRequest struct {
	UserUID string `form:"user_uid" binding:"required,min=1,alphanum"`
	ID      string `form:"id" binding:"required,min=1,alphanum"`
}

func (server *Server) getPastExperience(ctx *gin.Context) {
	var req getPastExperienceRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	pastExperience, err := server.esClient.GetPastExperience(ctx, req.UserUID, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if authUID != req.UserUID || err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}
	ctx.JSON(http.StatusOK, pastExperience)
}

type listPastExperiencesRequest struct {
	UserUID string `form:"user_uid" binding:"required,min=1,alphanum"`
}

func (server *Server) listPastExperiences(ctx *gin.Context) {
	var req listPastExperiencesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	pastExperiences, err := server.esClient.ListPastExperiences(ctx, req.UserUID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, pastExperiences)
}

type updatePastExperienceRequest struct {
	UserUID     string    `json:"user_uid" binding:"required,min=1,alphanum"`
	ID          string    `json:"id" binding:"required,min=1,alphanum"`
	Industry    string    `json:"industry,omitempty"`
	Employer    string    `json:"employer,omitempty"`
	JobTitle    string    `json:"job_title,omitempty"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	Present     bool      `json:"present"`
	Description string    `json:"description,omitempty"`
}

func (server *Server) updatePastExperience(ctx *gin.Context) {
	var req updatePastExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if authUID != req.UserUID || err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}
	pastExperience := es.PastExperience{
		ID:          req.ID,
		Industry:    req.Industry,
		Employer:    req.Employer,
		JobTitle:    req.JobTitle,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Present:     req.Present,
		Description: req.Description,
	}
	payload := &worker.PayloadPastExperience{
		PastExperience: pastExperience,
		UserUID:        req.UserUID,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.ProcessIn(5 * time.Second),
		asynq.Queue(worker.QueueDefault),
	}
	err = server.taskDistributor.DistributeTaskUpdatePastExperience(ctx, payload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to distribute task: %w", err)))
		return
	}
	ctx.JSON(http.StatusOK, statusResponse("update past experience task queued successfully"))
}

type deletePastExperienceRequest struct {
	UserUID string `json:"user_uid" binding:"required,min=1"`
	ID      string `json:"id" binding:"required,min=1"`
}

func (server *Server) deletePastExperience(ctx *gin.Context) {
	var req deletePastExperienceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authUID, err := firebase.GetUIDFromClaims(ctx)
	if authUID != req.UserUID || err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}
	payload := &worker.PayloadDeletePastExperience{
		PastExperienceID: req.ID,
		UserUID:          req.UserUID,
	}
	opts := []asynq.Option{
		asynq.MaxRetry(5),
		asynq.ProcessIn(2 * time.Second),
		asynq.Queue(worker.QueueCritical),
	}
	err = server.taskDistributor.DistributeTaskDeletePastExperience(ctx, payload, opts...)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(fmt.Errorf("failed to distribute task: %w", err)))
		return
	}

	ctx.JSON(http.StatusOK, statusResponse("delete past experience task queued successfully"))
}
