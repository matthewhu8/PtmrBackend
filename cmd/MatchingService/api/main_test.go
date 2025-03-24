package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/google"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, esClient elasticsearch.ESClient, gapi google.GAPI) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	require.NoError(t, err)
	server := NewServer(config, esClient, tokenMaker, gapi)
	server.SetupRouter()
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
