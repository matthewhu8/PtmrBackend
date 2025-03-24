package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store, esClient elasticsearch.ESClient, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	require.NoError(t, err)
	server := NewServer(config, store, esClient, tokenMaker, taskDistributor)
	server.SetupRouter()
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
