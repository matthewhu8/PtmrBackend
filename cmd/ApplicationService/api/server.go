package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
)

type Server struct {
	config          util.Config
	store           db.Store
	esClient        elasticsearch.ESClient
	router          *gin.Engine
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, esClient elasticsearch.ESClient, tokenMaker token.Maker, taskDistributor worker.TaskDistributor) *Server {
	return &Server{
		config:          config,
		store:           store,
		esClient:        esClient,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
	}
}

func (server *Server) SetupRouter() {
	router := gin.Default()
	authRoutes := router.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))

	// Candidate Application Routes
	authRoutes.POST("/candidate_applications", func(ctx *gin.Context) {
		server.createApplication(ctx, false)
	})
	authRoutes.GET("/candidate_applications/:employer_id", func(ctx *gin.Context) {
		server.getApplications(ctx, true)
	})
	authRoutes.PATCH("/candidate_applications/:candidate_id/:employer_id", func(ctx *gin.Context) {
		server.updateApplication(ctx, false)
	})
	authRoutes.DELETE("/candidate_applications/:candidate_username", func(ctx *gin.Context) {
		server.deleteApplication(ctx, false)
	})

	// Employer Application Routes
	authRoutes.POST("/employer_applications", func(ctx *gin.Context) {
		server.createApplication(ctx, true)
	})
	authRoutes.GET("/employer_applications/:candidate_id", func(ctx *gin.Context) {
		server.getApplications(ctx, false)
	})
	authRoutes.PATCH("/employer_applications/:employer_id/:candidate_id", func(ctx *gin.Context) {
		server.updateApplication(ctx, true)
	})
	authRoutes.DELETE("/employer_applications/:employer_username", func(ctx *gin.Context) {
		server.deleteApplication(ctx, true)
	})

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
