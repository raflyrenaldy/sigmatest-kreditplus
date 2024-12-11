// aws/s3/s3.go
package s3

import (
	"bytes"
	"context"
	"customer/sigmatech/app/constants"
	"customer/sigmatech/app/service/logger"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type IS3Client interface {
	ListBuckets() ([]string, error)
	ListObjects() ([]string, error)
	GetObject(objectName string) (*s3.GetObjectOutput, error)
	GetObjectSize(objectName string) (float64, error)
	PutObject(objectName string, body []byte) (*s3.PutObjectOutput, error)
	DeleteObject(objectName string) (*s3.DeleteObjectOutput, error)
	DeleteBucket() (*s3.DeleteBucketOutput, error)
	CreateBucket() (*s3.CreateBucketOutput, error)
	CopyObject(sourceBucketName, sourceObjectName, destinationBucketName, destinationObjectName string) (*s3.CopyObjectOutput, error)
}

// S3Service is a struct to interact with Amazon S3.
type S3Service struct {
	Client *s3.Client
}

var bucketName string

// NewS3Service initializes a new S3 service.
func NewS3Service() *S3Service {
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:       "aws",
			URL:               constants.Config.AWSConfig.S3Endpoint,
			HostnameImmutable: true,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(constants.Config.AWSConfig.AccessKeyId, constants.Config.AWSConfig.SecretAccessKey, "")),
	)
	if err != nil {
		logger.Logger(context.Background()).Fatalf("S3 service failed with error: %v", err)
	}

	bucketName = constants.Config.AWSConfig.S3BucketName

	return &S3Service{
		Client: s3.NewFromConfig(cfg),
	}
}

// ListBuckets returns a list of S3 bucket names.
func (s *S3Service) ListBuckets() ([]string, error) {
	input := &s3.ListBucketsInput{}
	resp, err := s.Client.ListBuckets(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var bucketNames []string
	for _, b := range resp.Buckets {
		bucketNames = append(bucketNames, aws.ToString(b.Name))
	}

	return bucketNames, nil
}

// ListObjects returns a list of S3 object names in a bucket.
func (s *S3Service) ListObjects() ([]string, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := s.Client.ListObjects(context.Background(), input)
	if err != nil {
		return nil, err
	}

	var objectNames []string
	for _, o := range resp.Contents {
		objectNames = append(objectNames, aws.ToString(o.Key))
	}

	return objectNames, nil
}

// GetObject returns an S3 object.
func (s *S3Service) GetObject(objectName string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}
	resp, err := s.Client.GetObject(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetObject returns an S3 object.
func (s *S3Service) GetObjectSize(objectName string) (float64, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}
	resp, err := s.Client.GetObject(context.Background(), input)
	if err != nil {
		return 0.00, nil
	}

	if resp.ContentLength == 0 {
		return 0.00, nil
	}

	return float64(resp.ContentLength), nil
}

// PutObject uploads an S3 object.
func (s *S3Service) PutObject(objectName string, body []byte) (*s3.PutObjectOutput, error) {
	// get the content type of the file
	contentType := http.DetectContentType(body)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		Body:        bytes.NewReader(body),
		ContentType: &contentType,
	}
	resp, err := s.Client.PutObject(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteObject deletes an S3 object.
func (s *S3Service) DeleteObject(objectName string) (*s3.DeleteObjectOutput, error) {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	}
	resp, err := s.Client.DeleteObject(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteBucket deletes an S3 bucket.
func (s *S3Service) DeleteBucket() (*s3.DeleteBucketOutput, error) {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := s.Client.DeleteBucket(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateBucket creates an S3 bucket.
func (s *S3Service) CreateBucket() (*s3.CreateBucketOutput, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	resp, err := s.Client.CreateBucket(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CopyObject copies an S3 object.
func (s *S3Service) CopyObject(sourceBucketName string, sourceObjectName string, destinationBucketName string, destinationObjectName string) (*s3.CopyObjectOutput, error) {
	input := &s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucketName),
		CopySource: aws.String(sourceBucketName + "/" + sourceObjectName),
		Key:        aws.String(destinationObjectName),
	}
	resp, err := s.Client.CopyObject(context.Background(), input)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
