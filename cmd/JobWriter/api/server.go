package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/google"
	"github.com/hankimmy/PtmrBackend/pkg/middleware"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
)

type Server struct {
	config     util.Config
	esClient   elasticsearch.ESClient
	router     *gin.Engine
	tokenMaker token.Maker
	gapi       google.GAPI
}

func NewServer(config util.Config, esClient elasticsearch.ESClient, tokenMaker token.Maker, gapi google.GAPI) *Server {
	return &Server{
		config:     config,
		esClient:   esClient,
		tokenMaker: tokenMaker,
		gapi:       gapi,
	}
}

func (server *Server) SetupRouter() {
	router := gin.Default()
	authRoutes := router.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))
	authRoutes.POST("/jobs/:employer_id", server.CreateJob)
	authRoutes.GET("/jobs/:job_id", server.GetJob)
	authRoutes.PATCH("/jobs/:job_id", server.UpdateJob)
	authRoutes.DELETE("/jobs/:job_id", server.DeleteJob)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
