package elasticsearch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSearchJobs(t *testing.T) {
	job1 := RandomJob(1)
	job1.Industry = "Tech"
	job1.EmploymentType = "Full-time"
	job1.Title = "Software Engineer"
	err := esClient.IndexJob(&job1)
	require.NoError(t, err)

	job2 := RandomJob(2)
	job2.Industry = "Tech"
	job2.EmploymentType = "Part-time"
	job2.Title = "Backend Developer"
	err = esClient.IndexJob(&job2)
	require.NoError(t, err)

	job3 := RandomJob(3)
	job3.Industry = "Healthcare"
	job3.EmploymentType = "Full-time"
	job3.Title = "Nurse"
	err = esClient.IndexJob(&job3)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	industry := job1.Industry
	employmentType := job1.EmploymentType
	title := job1.Title
	candidateLocation := job1.PreciseLocation // Use job1's location as candidate location
	distance := "10mi"                        // Arbitrary distance for the test

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)
	require.NotNil(t, jobs)

	require.Len(t, jobs, 1)

	expectedJob := job1
	actualJob := jobs[0]
	require.Equal(t, expectedJob.ID, actualJob.ID)
	require.Equal(t, expectedJob.Industry, actualJob.Industry)
	require.Equal(t, expectedJob.EmploymentType, actualJob.EmploymentType)
	require.Equal(t, expectedJob.Title, actualJob.Title)
	clearIndex(JobIdx)
}

func TestSearchJobs_MultipleMatches(t *testing.T) {
	job1 := RandomJob(1)
	job1.Industry = "Tech"
	job1.EmploymentType = "Full-time"
	job1.Title = "Software Engineer"
	err := esClient.IndexJob(&job1)
	require.NoError(t, err)

	job2 := RandomJob(2)
	job2.Industry = "Tech"
	job2.EmploymentType = "Full-time"
	job2.Title = "Software Engineer"
	err = esClient.IndexJob(&job2)
	require.NoError(t, err)

	job3 := RandomJob(3)
	job3.Industry = "Tech"
	job3.EmploymentType = "Full-time"
	job3.Title = "Software Engineer"
	err = esClient.IndexJob(&job3)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	industry := "Tech"
	employmentType := "Full-time"
	title := "Software Engineer"
	candidateLocation := job1.PreciseLocation
	distance := "10mi"

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)
	require.NotNil(t, jobs)

	require.Len(t, jobs, 3)

	jobIDs := []string{job1.ID, job2.ID, job3.ID}
	for _, job := range jobs {
		require.Contains(t, jobIDs, job.ID)
		require.Equal(t, industry, job.Industry)
		require.Equal(t, employmentType, job.EmploymentType)
		require.Equal(t, title, job.Title)
	}

	clearIndex(JobIdx)
}

func TestSearchJobs_NoMatches(t *testing.T) {
	job := RandomJob(1)
	job.Industry = "Healthcare"
	job.EmploymentType = "Part-time"
	job.Title = "Nurse"
	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	industry := "Tech"
	employmentType := "Full-time"
	title := "Software Engineer"
	candidateLocation := job.PreciseLocation
	distance := "10mi"

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)
	require.Empty(t, jobs)
	clearIndex(JobIdx)
}

func TestSearchJobs_PartialMatches(t *testing.T) {
	job := RandomJob(1)
	job.Industry = "Tech"
	job.EmploymentType = "Part-time"
	job.Title = "Backend Developer"
	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	industry := "Tech"
	employmentType := "Full-time"
	title := "Backend Developer"
	candidateLocation := job.PreciseLocation
	distance := "10mi"

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)

	require.Empty(t, jobs)
	clearIndex(JobIdx)
}

func TestSearchJobs_LargeDataset(t *testing.T) {
	var expectedJobs []Job
	var candidateLocation GeoPoint
	for i := int64(0); i < 1000; i++ {
		job := RandomJob(i)
		if i%100 == 0 {
			job.Industry = "Tech"
			job.EmploymentType = "Full-time"
			job.Title = "Software Engineer"
			expectedJobs = append(expectedJobs, job)
			candidateLocation = job.PreciseLocation // Use the first matching job's location
		}
		err := esClient.IndexJob(&job)
		require.NoError(t, err)
	}

	time.Sleep(5 * time.Second)

	industry := "Tech"
	employmentType := "Full-time"
	title := "Software Engineer"
	distance := "10mi"

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)

	require.NotNil(t, jobs)
	require.LessOrEqual(t, len(jobs), 20, "Number of returned jobs should be less than or equal to 20")

	for _, job := range jobs {
		require.Equal(t, industry, job.Industry)
		require.Equal(t, employmentType, job.EmploymentType)
		require.Equal(t, title, job.Title)
	}

	clearIndex(JobIdx)
}

func TestSearchJobs_FuzzyMatching(t *testing.T) {
	job := RandomJob(1)
	job.Industry = "Tech"
	job.EmploymentType = "Full-time"
	job.Title = "Senior Software Engineer"
	err := esClient.IndexJob(&job)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	industry := "Tech"
	employmentType := "Full-time"
	title := "Senior Softwre Enginer"
	candidateLocation := job.PreciseLocation
	distance := "10mi"

	jobs, err := esClient.SearchJobs(industry, employmentType, title, distance, candidateLocation)
	require.NoError(t, err)

	require.Len(t, jobs, 1)
	require.Equal(t, job.ID, jobs[0].ID)
	require.Equal(t, job.Industry, jobs[0].Industry)
	require.Equal(t, job.EmploymentType, jobs[0].EmploymentType)
	require.Equal(t, job.Title, jobs[0].Title)

	clearIndex(JobIdx)
}
