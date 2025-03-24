package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	fb "firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
	"github.com/hankimmy/PtmrBackend/pkg/firebase"
	mockfb "github.com/hankimmy/PtmrBackend/pkg/firebase/mock"
	"github.com/hankimmy/PtmrBackend/pkg/token"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/hankimmy/PtmrBackend/pkg/worker"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor, esClient elasticsearch.ESClient, authClient firebase.AuthClientFirebase, rateLimiter *RateLimiter) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	require.NoError(t, err)
	server := NewServer(config, store, esClient, taskDistributor, tokenMaker, authClient, rateLimiter)
	server.SetupRouter()
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func marshalRequestBody(t *testing.T, body gin.H) []byte {
	data, err := json.Marshal(body)
	require.NoError(t, err)
	return data
}

func createNewRequest(t *testing.T, method, url string, body []byte) *http.Request {
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	require.NoError(t, err)
	return request
}

func mockCandidateMiddleware(t *testing.T, uid string) *mockfb.MockAuthClientFirebase {
	authCtrl := gomock.NewController(t)
	auth := mockfb.NewMockAuthClientFirebase(authCtrl)

	auth.EXPECT().VerifyIDToken(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(&fb.Token{
			UID: uid,
			Claims: map[string]interface{}{
				"uid":  uid,
				"role": string(db.RoleCandidate),
			},
		}, nil)

	return auth
}
