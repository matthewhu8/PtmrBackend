package s3

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
)

func TestListS3Objects(t *testing.T) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")

	sess, err := NewSession(region, accessKey, secretKey)
	assert.NoError(t, err)

	client := NewS3Client(sess)

	// Act
	keys, err := client.ListS3Objects(bucket)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, keys, "The bucket should contain at least one object")
}

func TestGetJSONFromS3(t *testing.T) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")

	key := "data.json"

	sess, err := NewSession(region, accessKey, secretKey)
	assert.NoError(t, err)

	client := NewS3Client(sess)

	jobs, err := client.GetJSONFromS3(bucket, key)

	assert.NoError(t, err)
	assert.NotNil(t, jobs)
	assert.NotEmpty(t, jobs)

	expectedJob := elasticsearch.Job{
		HiringOrganization: "Tech Innovators Inc.",
		Title:              "Software Engineer",
		Industry:           "Technology",
		JobLocation:        "San Francisco, CA",
		DatePosted:         "2024-08-15",
		Description:        "Develop and maintain web applications using modern frameworks. Collaborate with cross-functional teams to deliver high-quality software.",
		EmploymentType:     "Full-time",
		DestinationURL:     "https://techinnovators.com/jobs/software-engineer",
	}

	found := false
	for _, job := range jobs {
		if job.HiringOrganization == expectedJob.HiringOrganization &&
			job.Title == expectedJob.Title &&
			job.Industry == expectedJob.Industry &&
			job.JobLocation == expectedJob.JobLocation &&
			job.DatePosted == expectedJob.DatePosted &&
			job.Description == expectedJob.Description &&
			job.EmploymentType == expectedJob.EmploymentType &&
			job.DestinationURL == expectedJob.DestinationURL {
			found = true
			break
		}
	}

	assert.True(t, found, "Expected job not found in the retrieved jobs")
	assert.GreaterOrEqual(t, len(jobs), 10, "Expected at least 10 jobs in the JSON")
}
