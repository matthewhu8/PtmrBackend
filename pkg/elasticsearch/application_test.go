package elasticsearch

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/require"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func TestIndexApplication(t *testing.T) {
	docID, application := makeRandomCandidateApp(t)

	// Verify document was indexed
	res, err := esClient.Client.Get().
		Index(CandidateAppIdx).
		Id(docID).
		Do(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res.Source)

	var indexedApplication map[string]interface{}
	err = json.Unmarshal(res.Source, &indexedApplication)
	require.NoError(t, err)
	require.Equal(t, application, indexedApplication)
}

func TestGetApplication(t *testing.T) {
	docID, application := makeRandomCandidateApp(t)
	gotApplication, err := esClient.GetCandidateApplication(context.Background(), docID)
	require.NoError(t, err)
	require.Equal(t, application, gotApplication)
}

func TestUpdateApplication(t *testing.T) {
	docID, _ := makeRandomCandidateApp(t)

	// Update the application
	updatedApplication := map[string]interface{}{
		"id":    "test_application_3",
		"field": "new_value",
	}
	err := esClient.UpdateCandidateApplication(context.Background(), docID, updatedApplication)
	require.NoError(t, err)

	// Verify the application was updated
	gotApplication, err := esClient.GetCandidateApplication(context.Background(), docID)
	require.NoError(t, err)
	require.Equal(t, updatedApplication, gotApplication)
}

func TestDeleteApplication(t *testing.T) {
	docID, _ := makeRandomCandidateApp(t)
	err := esClient.DeleteCandidateApplication(context.Background(), docID)
	require.NoError(t, err)

	// Verify the application was deleted
	_, err = esClient.Client.Get().
		Index(CandidateAppIdx).
		Id(docID).
		Do(context.Background())
	require.Error(t, err)
	require.True(t, elastic.IsNotFound(err))
}

func makeRandomCandidateApp(t *testing.T) (string, map[string]interface{}) {
	docID := util.RandomString(5)
	application := map[string]interface{}{
		"field": util.RandomString(10),
	}

	err := esClient.IndexCandidateApplication(context.Background(), docID, application)
	require.NoError(t, err)
	return docID, application
}

func TestIndexEmployerApplication(t *testing.T) {
	docID, application := makeRandomEmployerApp(t)

	// Verify document was indexed
	res, err := esClient.Client.Get().
		Index(EmployerAppIdx).
		Id(docID).
		Do(context.Background())
	require.NoError(t, err)
	require.NotNil(t, res.Source)

	var indexedApplication map[string]interface{}
	err = json.Unmarshal(res.Source, &indexedApplication)
	require.NoError(t, err)
	require.Equal(t, application, indexedApplication)
}

func TestGetEmployerApplication(t *testing.T) {
	docID, application := makeRandomEmployerApp(t)
	gotApplication, err := esClient.GetEmployerApplication(context.Background(), docID)
	require.NoError(t, err)
	require.Equal(t, application, gotApplication)
}

func TestUpdateEmployerApplication(t *testing.T) {
	docID, _ := makeRandomEmployerApp(t)

	// Update the application
	updatedApplication := map[string]interface{}{
		"id":    "test_application_3",
		"field": "new_value",
	}
	err := esClient.UpdateEmployerApplication(context.Background(), docID, updatedApplication)
	require.NoError(t, err)

	// Verify the application was updated
	gotApplication, err := esClient.GetEmployerApplication(context.Background(), docID)
	require.NoError(t, err)
	require.Equal(t, updatedApplication, gotApplication)
}

func TestDeleteEmployerApplication(t *testing.T) {
	docID, _ := makeRandomEmployerApp(t)
	err := esClient.DeleteEmployerApplication(context.Background(), docID)
	require.NoError(t, err)

	// Verify the application was deleted
	_, err = esClient.Client.Get().
		Index(EmployerAppIdx).
		Id(docID).
		Do(context.Background())
	require.Error(t, err)
	require.True(t, elastic.IsNotFound(err))
}

func makeRandomEmployerApp(t *testing.T) (string, map[string]interface{}) {
	docID := util.RandomString(5)
	application := map[string]interface{}{
		"field": util.RandomString(10),
	}

	err := esClient.IndexEmployerApplication(context.Background(), docID, application)
	require.NoError(t, err)
	return docID, application
}
