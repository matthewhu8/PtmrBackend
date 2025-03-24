package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
)

type getCandidateBatchFeedRequest struct {
	CandidateID    int64  `uri:"candidate_id" binding:"required,min=1"`
	Industry       string `json:"industry"`
	EmploymentType string `json:"employment_type"`
	Title          string `json:"title"`
	Location       string `json:"location"`
	Distance       string `json:"distance"`
}

func (server *Server) GetCandidateBatchFeed(ctx *gin.Context) {
	var req getCandidateBatchFeedRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if authPayload.Role != db.RoleCandidate || authPayload.RoleID != req.CandidateID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("not authorized to access this resource")))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	lat, lon, err := server.gapi.GetLatLon(req.Location)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	candidateLocation := elasticsearch.GeoPoint{
		Lat: lat,
		Lon: lon,
	}

	jobs, err := server.esClient.SearchJobs(req.Industry, req.EmploymentType, req.Title, req.Distance, candidateLocation)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if jobs == nil {
		jobs = []elasticsearch.Job{}
	}

	ctx.JSON(http.StatusOK, jobs)
}
