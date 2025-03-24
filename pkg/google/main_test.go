package google

import (
	"os"
	"testing"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func TestMain(m *testing.M) {
	config, _ := util.LoadConfig(".")
	os.Setenv("GOOGLE_API_KEY", config.GoogleAPIKey)
	os.Exit(m.Run())
}

func skipAPICalls() bool {
	return os.Getenv("SKIP_API_CALLS") == "true"
}
