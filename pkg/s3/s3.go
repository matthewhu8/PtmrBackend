package s3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hankimmy/PtmrBackend/pkg/elasticsearch"
)

type Client interface {
	ListS3Objects(bucket string) ([]string, error)
	GetJSONFromS3(bucket, key string) ([]elasticsearch.Job, error)
}

type s3Client struct {
	sess *session.Session
}

func NewSession(region, accessKey, secretKey string) (*session.Session, error) {
	config := &aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	return sess, nil
}

func NewS3Client(sess *session.Session) Client {
	return &s3Client{sess: sess}
}

func (c *s3Client) ListS3Objects(bucket string) ([]string, error) {
	svc := s3.New(c.sess)
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in S3: %v", err)
	}

	var keys []string
	for _, item := range result.Contents {
		keys = append(keys, *item.Key)
	}

	return keys, nil
}

func (c *s3Client) GetJSONFromS3(bucket, key string) ([]elasticsearch.Job, error) {
	svc := s3.New(c.sess)
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer result.Body.Close()

	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %v", err)
	}

	var jobs []elasticsearch.Job
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return jobs, nil
}
