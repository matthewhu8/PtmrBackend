package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/service"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
)

type createApplicationRequest struct {
	CandidateID       int64                `json:"candidate_id,omitempty"`
	EmployerID        int64                `json:"employer_id,omitempty"`
	ApplicationStatus db.ApplicationStatus `json:"application_status" binding:"required"`
	ApplicationDoc    json.RawMessage      `json:"application_doc" binding:"required"`
	JobDocID          string               `json:"job_doc_id,omitempty"` // Only for candidate applications
}

func (server *Server) createApplication(ctx *gin.Context, isEmployer bool) {
	var req createApplicationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	if err := validateIDs(req.CandidateID, req.EmployerID); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	var docID string
	appDoc, err := convertApplicationDoc(req.ApplicationDoc)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}

	if isEmployer {
		if authPayload.Role != db.RoleEmployer || authPayload.RoleID != req.EmployerID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account is not employer or belong to user")))
			return
		}
		message, err := getMessageFromApplicationDoc(req.ApplicationDoc)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
			return
		}
		arg := db.CreateEmployerApplicationParams{
			EmployerID:        authPayload.RoleID,
			CandidateID:       req.CandidateID,
			Message:           message,
			ApplicationStatus: req.ApplicationStatus,
		}
		result, err := server.store.CreateEmployerApplication(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, result)
	} else {
		if authPayload.Role != db.RoleCandidate || authPayload.RoleID != req.CandidateID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account is not candidate or belong to user")))
			return
		}
		docID = fmt.Sprintf("%d_%d", authPayload.RoleID, req.EmployerID)
		arg := db.CreateCandidateApplicationTxParams{
			CreateCandidateApplicationParams: db.CreateCandidateApplicationParams{
				CandidateID:        authPayload.RoleID,
				EmployerID:         req.EmployerID,
				ElasticsearchDocID: docID,
				JobDocID:           req.JobDocID,
				ApplicationStatus:  req.ApplicationStatus,
			},
			AppDoc:      appDoc,
			AfterCreate: server.afterCandidateCreateApp(ctx),
		}
		result, err := server.store.CreateCandidateApplicationTx(ctx, arg)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
			return
		}
		ctx.JSON(http.StatusOK, result.CandidateApplication)
	}
}

type getApplicationsRequest struct {
	CandidateID int64 `uri:"candidate_id,omitempty"`
	EmployerID  int64 `uri:"employer_id,omitempty"`
}

func (server *Server) getApplications(ctx *gin.Context, isEmployer bool) {
	var req getApplicationsRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}

	var err error
	var res []map[string]interface{}
	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)

	if !isEmployer {
		var applications []db.EmployerApplication
		applications, err = server.store.GetEmployerApplicationsByCandidate(ctx, req.CandidateID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
			return
		}

		// Authorization check
		if len(applications) > 0 && applications[0].CandidateID != authPayload.RoleID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account doesn't belong to the authenticated user")))
			return
		}

		ctx.JSON(http.StatusOK, applications)
	} else {
		var applications []db.CandidateApplication
		applications, err = server.store.GetCandidateApplicationsByEmployer(ctx, req.EmployerID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
			return
		}

		// Authorization check
		if len(applications) > 0 && applications[0].EmployerID != authPayload.RoleID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account doesn't belong to the authenticated user")))
			return
		}

		// Fetching Elasticsearch documents
		for _, application := range applications {
			esResult, err := server.esClient.GetCandidateApplication(ctx, application.ElasticsearchDocID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(fmt.Errorf("failed to get application in Elasticsearch: %v", err)))
				return
			}
			res = append(res, map[string]interface{}{
				"metadata": application,
				"document": esResult,
			})
		}
		ctx.JSON(http.StatusOK, res)
	}
}

type updateApplicationRequest struct {
	CandidateID       int64                `uri:"candidate_id,omitempty"`
	EmployerID        int64                `uri:"employer_id,omitempty"`
	ApplicationStatus db.ApplicationStatus `json:"application_status"`
	ApplicationDoc    json.RawMessage      `json:"application_doc"`
}

func (req *updateApplicationRequest) validateIDs() error {
	if req.CandidateID == 0 || req.EmployerID == 0 {
		return errors.New("either candidate_id or employer_id must be provided")
	}
	return nil
}

func (server *Server) updateApplication(ctx *gin.Context, isEmployer bool) {
	var req updateApplicationRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}

	if err := req.validateIDs(); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if authPayload.RoleID != req.CandidateID && authPayload.RoleID != req.EmployerID {
		ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account doesn't belong to the authenticated user")))
		return
	}
	var docID string
	var arg interface{}
	var application interface{}
	var err error

	if isEmployer {
		docID = fmt.Sprintf("%d_%d", req.EmployerID, req.CandidateID)
		arg = db.UpdateEmployerApplicationParams{
			CandidateID: req.CandidateID,
			EmployerID:  req.EmployerID,
			ApplicationStatus: db.NullApplicationStatus{
				ApplicationStatus: req.ApplicationStatus,
				Valid:             req.ApplicationStatus != "",
			},
		}
		application, err = server.store.UpdateEmployerApplication(ctx, arg.(db.UpdateEmployerApplicationParams))
	} else {
		docID = fmt.Sprintf("%d_%d", req.CandidateID, req.EmployerID)
		arg = db.UpdateCandidateApplicationParams{
			CandidateID: req.CandidateID,
			EmployerID:  req.EmployerID,
			ApplicationStatus: db.NullApplicationStatus{
				ApplicationStatus: req.ApplicationStatus,
				Valid:             req.ApplicationStatus != "",
			},
			ElasticsearchDocID: pgtype.Text{
				String: docID,
				Valid:  true,
			},
		}
		application, err = server.store.UpdateCandidateApplication(ctx, arg.(db.UpdateCandidateApplicationParams))
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
		return
	}

	if req.ApplicationDoc != nil && !isEmployer {
		var applicationDoc map[string]interface{}
		if err := json.Unmarshal(req.ApplicationDoc, &applicationDoc); err != nil {
			ctx.JSON(http.StatusBadRequest, service.ErrorResponse(fmt.Errorf("invalid application_doc: %v", err)))
			return
		}
		if err := server.esClient.UpdateCandidateApplication(ctx, docID, applicationDoc); err != nil {
			ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(fmt.Errorf("failed to update application in Elasticsearch: %v", err)))
			return
		}
	}

	ctx.JSON(http.StatusOK, application)
}

type deleteApplicationRequest struct {
	CandidateUsername string `uri:"candidate_username,omitempty""`
	EmployerUsername  string `uri:"employer_username,omitempty""`
	CandidateID       int64  `json:"candidate_id,omitempty"`
	EmployerID        int64  `json:"employer_id,omitempty"`
}

func (server *Server) deleteApplication(ctx *gin.Context, isEmployer bool) {
	var req deleteApplicationRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	if err := validateIDs(req.CandidateID, req.EmployerID); err != nil {
		ctx.JSON(http.StatusBadRequest, service.ErrorResponse(err))
		return
	}
	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	var docID string
	var err error
	if isEmployer {
		if authPayload.Role != db.RoleEmployer || authPayload.RoleID != req.EmployerID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account is not employer or belong to user")))
			return
		}
		docID = fmt.Sprintf("%d_%d", authPayload.RoleID, req.CandidateID)
		arg := db.DeleteApplicationTxParams{
			DeleteEmployerApplicationParams: db.DeleteEmployerApplicationParams{
				CandidateID: req.CandidateID,
				EmployerID:  authPayload.RoleID,
			},
			IsEmployer:  isEmployer,
			DocID:       docID,
			AfterDelete: server.afterEmployerDeleteApp(ctx),
		}
		err = server.store.DeleteApplicationTx(ctx, arg)
	} else {
		if authPayload.Role != db.RoleCandidate || authPayload.RoleID != req.CandidateID {
			ctx.JSON(http.StatusUnauthorized, service.ErrorResponse(errors.New("account is not employer or belong to user")))
			return
		}
		docID = fmt.Sprintf("%d_%d", authPayload.RoleID, req.EmployerID)
		arg := db.DeleteApplicationTxParams{
			DeleteCandidateApplicationParams: db.DeleteCandidateApplicationParams{
				CandidateID: authPayload.RoleID,
				EmployerID:  req.EmployerID,
			},
			IsEmployer:  isEmployer,
			DocID:       docID,
			AfterDelete: server.afterCandidateDeleteApp(ctx),
		}
		err = server.store.DeleteApplicationTx(ctx, arg)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, service.ErrorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, service.StatusResponse("deleted"))
}

func (server *Server) afterCandidateCreateApp(ctx *gin.Context) func(docID string, appDoc map[string]interface{}) error {
	return server.enqueueCreateAppTask(worker.TaskCreateCandidateApp, ctx)
}

func (server *Server) afterEmployerCreateApp(ctx *gin.Context) func(docID string, appDoc map[string]interface{}) error {
	return server.enqueueCreateAppTask(worker.TaskCreateEmployerApp, ctx)
}

func (server *Server) enqueueCreateAppTask(taskType string, ctx *gin.Context) func(docID string, appDoc map[string]interface{}) error {
	return func(docID string, appDoc map[string]interface{}) error {
		taskPayload := &worker.PayloadCreateApplication{
			AppDoc: appDoc,
			DocID:  docID,
		}
		opts := []asynq.Option{
			asynq.MaxRetry(10),
			asynq.ProcessIn(10 * time.Second),
			asynq.Queue(worker.QueueCritical),
		}

		switch taskType {
		case worker.TaskCreateCandidateApp:
			return server.taskDistributor.DistributeTaskCreateCandidateApplication(ctx, taskPayload, opts...)
		case worker.TaskCreateEmployerApp:
			return server.taskDistributor.DistributeTaskCreateEmployerApplication(ctx, taskPayload, opts...)
		default:
			return errors.New("unsupported task type")
		}
	}
}

func (server *Server) afterCandidateDeleteApp(ctx *gin.Context) func(docID string) error {
	return server.enqueueDeleteAppTask(worker.TaskDeleteCandidateApp, ctx)
}

func (server *Server) afterEmployerDeleteApp(ctx *gin.Context) func(docID string) error {
	return server.enqueueDeleteAppTask(worker.TaskDeleteEmployerApp, ctx)
}

func (server *Server) enqueueDeleteAppTask(taskType string, ctx *gin.Context) func(docID string) error {
	return func(docID string) error {
		taskPayload := &worker.PayloadDeleteApplication{
			DocID: docID,
		}
		opts := []asynq.Option{
			asynq.MaxRetry(10),
			asynq.ProcessIn(10 * time.Second),
			asynq.Queue(worker.QueueCritical),
		}

		switch taskType {
		case worker.TaskDeleteCandidateApp:
			return server.taskDistributor.DistributeTaskDeleteCandidateApplication(ctx, taskPayload, opts...)
		case worker.TaskDeleteEmployerApp:
			return server.taskDistributor.DistributeTaskDeleteEmployerApplication(ctx, taskPayload, opts...)
		default:
			return errors.New("unsupported task type")
		}
	}
}

// validateIDs checks that at least one ID (CandidateID or EmployerID) is provided
func validateIDs(candidateID, employerID int64) error {
	if candidateID == 0 || employerID == 0 {
		return errors.New("either candidate_id or employer_id must be provided")
	}
	return nil
}

func convertApplicationDoc(app json.RawMessage) (map[string]interface{}, error) {
	var applicationDoc map[string]interface{}
	if err := json.Unmarshal(app, &applicationDoc); err != nil {
		return nil, fmt.Errorf("invalid application_doc: %v", err)
	}
	return applicationDoc, nil
}

func getMessageFromApplicationDoc(doc json.RawMessage) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(doc, &obj)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal ApplicationDoc: %w", err)
	}
	message, ok := obj["message"].(string)
	if !ok {
		return "", fmt.Errorf("message key not found or is not a string")
	}
	return message, nil
}
