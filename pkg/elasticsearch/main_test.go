package elasticsearch

import (
	"context"
	"os"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog/log"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

var esClient *ESClientImpl

func newTestESClient(config util.Config) *ESClientImpl {
	client, err := elastic.NewClient(elastic.SetURL(config.ESSource), elastic.SetSniff(false))
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config:")
	}
	return &ESClientImpl{Client: client}
}

func TestMain(m *testing.M) {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config:")
	}

	esClient = newTestESClient(config)
	code := m.Run()
	clearIndex(CandidateIdx)
	clearIndex(JobIdx)
	clearIndex(CandidateAppIdx)
	os.Exit(code)

}
func clearIndex(indexName string) {
	_, err := esClient.Client.DeleteByQuery().
		Index(indexName).
		Body(`{"query": {"match_all": {}}}`).
		Do(context.Background())
	if err != nil {
		log.Err(err)
	}
}
