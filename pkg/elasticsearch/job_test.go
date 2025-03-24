package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndexJob(t *testing.T) {
	job := RandomJob(1)
	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	res, err := esClient.Client.Get().
		Index(JobIdx).
		Id(job.ID).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)
	jobJSON, err := json.Marshal(source)
	require.NoError(t, err)

	var body bytes.Buffer
	body.Write(jobJSON)
	requireBodyMatchJob(t, &body, job)
}

func TestGetJob(t *testing.T) {
	job := RandomJob(1)

	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	retrievedJob, err := esClient.GetJob(job.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedJob)

	jobJSON, err := json.Marshal(retrievedJob)
	require.NoError(t, err)

	var body bytes.Buffer
	body.Write(jobJSON)
	requireBodyMatchJob(t, &body, *retrievedJob)
}

func TestUpdateJob(t *testing.T) {
	job := RandomJob(1)

	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	// Update job
	job.Title = "Senior Software Engineer"
	err = esClient.UpdateJob(job.ID, &job)
	require.NoError(t, err)

	// Verify job is updated
	res, err := esClient.Client.Get().
		Index(JobIdx).
		Id(job.ID).
		Do(context.Background())
	require.NoError(t, err)

	var source map[string]interface{}
	err = json.Unmarshal(res.Source, &source)
	require.NoError(t, err)
	jobJSON, err := json.Marshal(source)
	require.NoError(t, err)

	var body bytes.Buffer
	body.Write(jobJSON)
	requireBodyMatchJob(t, &body, job)
}

func TestDeleteJob(t *testing.T) {
	job := RandomJob(1)

	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	err = esClient.DeleteJob(job.ID)
	require.NoError(t, err)

	retrievedJob, err := esClient.GetJob(job.ID)
	require.NoError(t, err)
	require.Nil(t, retrievedJob)
}

func requireBodyMatchJob(t *testing.T, body *bytes.Buffer, job Job) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotJob Job
	err = json.Unmarshal(data, &gotJob)
	require.NoError(t, err)

	require.Equal(t, job.ID, gotJob.ID)
	require.Equal(t, job.Title, gotJob.Title)
	require.Equal(t, job.Description, gotJob.Description)
	require.Equal(t, job.HiringOrganization, gotJob.HiringOrganization)
	require.Equal(t, job.JobLocation, gotJob.JobLocation)
	require.Equal(t, job.EmploymentType, gotJob.EmploymentType)
	require.Equal(t, job.DestinationURL, gotJob.DestinationURL)
	require.Equal(t, job.IsUserCreated, gotJob.IsUserCreated)
}
