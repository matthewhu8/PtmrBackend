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
	authRoutes.GET("/feed/:candidate_id", server.GetCandidateBatchFeed)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
