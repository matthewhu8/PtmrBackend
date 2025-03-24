package s3

import (
	"os"
	"testing"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func TestMain(m *testing.M) {
	config, _ := util.LoadConfig(".")
	os.Setenv("AWS_REGION", config.AWSRegion)
	os.Setenv("S3_ACCESS_KEY", config.S3AccessKey)
	os.Setenv("S3_SECRET_KEY", config.S3SecretKey)
	os.Setenv("S3_BUCKET", config.S3Bucket)
	os.Exit(m.Run())
}
