package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
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
	auth            firebase.AuthClientFirebase
	rateLimiter     *RateLimiter
}

func NewServer(config util.Config, store db.Store, esClient elasticsearch.ESClient, taskDistributor worker.TaskDistributor, tokenMaker token.Maker, auth firebase.AuthClientFirebase, rateLimiter *RateLimiter) *Server {
	return &Server{
		config:          config,
		store:           store,
		esClient:        esClient,
		tokenMaker:      tokenMaker,
		taskDistributor: taskDistributor,
		auth:            auth,
		rateLimiter:     rateLimiter,
	}
}

func (server *Server) SetupRouter() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:9000"}, // Allow frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.POST("/users/", server.sendEmailVerification)
	router.POST("/resend-verification-email", server.resendEmailVerification)

	candidateRoutes := router.Group("/candidates").Use(firebase.AuthMiddleware(server.auth, string(db.RoleCandidate)))
	candidateRoutes.PUT("/", server.createCandidate)
	candidateRoutes.PATCH("/", server.updateCandidate)
	candidateRoutes.GET("/", server.getCandidate)
	candidateRoutes.DELETE("/", server.deleteCandidate)

	pastExpRoutes := router.Group("/candidates/past_experience").Use(firebase.AuthMiddleware(server.auth, string(db.RoleCandidate)))
	pastExpRoutes.PUT("/", server.createPastExperience)
	pastExpRoutes.GET("/", server.getPastExperience)
	pastExpRoutes.GET("/all", server.listPastExperiences)
	pastExpRoutes.PATCH("/", server.updatePastExperience)
	pastExpRoutes.DELETE("/", server.deletePastExperience)

	authRoutes := router.Group("/").Use(middleware.AuthMiddleware(server.tokenMaker))
	authRoutes.POST("/employers", server.createEmployer)
	authRoutes.GET("/employers/:id", server.getEmployer)
	authRoutes.GET("/employers", server.listEmployer)
	authRoutes.PATCH("/employers/:id", server.updateEmployer)

	server.router = router
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func statusResponse(msg string) gin.H {
	return gin.H{"status": msg}
}
