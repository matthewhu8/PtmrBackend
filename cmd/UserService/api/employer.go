package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/jackc/pgx/v5/pgtype"
)

type createEmployerRequest struct {
	BusinessName        string `json:"business_name" binding:"required"`
	BusinessEmail       string `json:"business_email" binding:"required"`
	BusinessPhone       string `json:"business_phone" binding:"required"`
	Location            string `json:"location"`
	Industry            string `json:"industry"`
	ProfilePhoto        string `json:"profile_photo"`
	BusinessDescription string `json:"business_description"`
}

func (server *Server) createEmployer(ctx *gin.Context) {
	var req createEmployerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)

	arg := db.CreateEmployerParams{
		Username:            authPayload.Username,
		BusinessName:        req.BusinessName,
		BusinessEmail:       req.BusinessEmail,
		BusinessPhone:       req.BusinessPhone,
		Location:            req.Location,
		Industry:            req.Industry,
		ProfilePhoto:        req.ProfilePhoto,
		BusinessDescription: req.BusinessDescription,
	}

	employer, err := server.store.CreateEmployer(ctx, arg)
	if err != nil {
		server.handleDatabaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, employer)
}

type getEmployerRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEmployer(ctx *gin.Context) {
	var req getEmployerRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	employer, err := server.store.GetEmployer(ctx, req.ID)
	if err != nil {
		server.handleDatabaseError(ctx, err)
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if employer.Username != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}

	ctx.JSON(http.StatusOK, employer)
}

type updateEmployerRequest struct {
	EmployerID          int64  `uri:"id" binding:"required,min=1"`
	BusinessName        string `json:"business_name"`
	BusinessEmail       string `json:"business_email"`
	BusinessPhone       string `json:"business_phone"`
	Location            string `json:"location"`
	Industry            string `json:"industry"`
	ProfilePhoto        string `json:"profile_photo"`
	BusinessDescription string `json:"business_description"`
}

func (server *Server) updateEmployer(ctx *gin.Context) {
	var req updateEmployerRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if authPayload.RoleID != req.EmployerID {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("account doesn't belong to the candidate user")))
	}
	arg := db.UpdateEmployerParams{
		ID: authPayload.RoleID,
		BusinessName: pgtype.Text{
			String: req.BusinessName,
			Valid:  req.BusinessName != "",
		},
		BusinessEmail: pgtype.Text{
			String: req.BusinessEmail,
			Valid:  req.BusinessEmail != "",
		},
		BusinessPhone: pgtype.Text{
			String: req.BusinessPhone,
			Valid:  req.BusinessPhone != "",
		},
		Location: pgtype.Text{
			String: req.Location,
			Valid:  req.Location != "",
		},
		Industry: pgtype.Text{
			String: req.Industry,
			Valid:  req.Industry != "",
		},
		ProfilePhoto: pgtype.Text{
			String: req.ProfilePhoto,
			Valid:  req.ProfilePhoto != "",
		},
		BusinessDescription: pgtype.Text{
			String: req.BusinessDescription,
			Valid:  req.BusinessDescription != "",
		},
	}

	employer, err := server.store.UpdateEmployer(ctx, arg)
	if err != nil {
		server.handleDatabaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, employer)
}

type listEmployerRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listEmployer(ctx *gin.Context) {
	var req listEmployerRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	arg := db.ListEmployersParams{
		Username: authPayload.Username,
		Limit:    req.PageSize,
		Offset:   (req.PageID - 1) * req.PageSize,
	}

	employers, err := server.store.ListEmployers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, employers)
}

func (server *Server) handleDatabaseError(ctx *gin.Context, err error) {
	if errors.Is(err, db.ErrRecordNotFound) {
		server.respondWithError(ctx, http.StatusNotFound, err)
		return
	}
	switch db.ErrorCode(err) {
	case db.UniqueViolation:
		server.respondWithError(ctx, http.StatusForbidden, err)
	case db.ForeignKeyViolation:
		server.respondWithError(ctx, http.StatusForbidden, err)
	default:
		server.respondWithError(ctx, http.StatusInternalServerError, err)
	}
}
