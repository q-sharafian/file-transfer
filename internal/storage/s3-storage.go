package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	l "github.com/q-sharafian/file-transfer/pkg/logger"
)

type S3Storage struct {
	// S3 client
	s3         *s3.Client
	presignS3  *s3.PresignClient
	bucketName string
	logger     l.Logger
}

func NewS3Storage(logger l.Logger) Storage {
	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))
	// svc := s3.New(sess, &aws.Config{
	// 	Region:   aws.String(os.Getenv("S3_REGION")),
	// 	Endpoint: aws.String(os.Getenv("S3_ENDPOINT")),
	// })
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("default")) //os.Getenv("S3_REGION")))

	if err != nil {
		logger.Panicf("Failed to init S3 storage service: %s", err.Error())
	}
	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
	})
	presignClient := s3.NewPresignClient(client)

	return &S3Storage{
		client,
		presignClient,
		os.Getenv("S3_BUCKET_NAME"),
		logger,
	}
}

func (s *S3Storage) UploadFile(fileInfo UploadFileInfo, expireTime time.Duration) (url.URL, error) {
	presignPutObject, err := s.presignS3.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:   &s.bucketName,
		Key:      aws.String(fmt.Sprintf("%s.%s", fileInfo.FileName, fileInfo.FileExtension.String())),
		Metadata: fileInfo.Metadata,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expireTime
	})

	if err != nil {
		return url.URL{}, fmt.Errorf("failed to create presign uploading link with key name %s: %s",
			fileInfo.FileName, err.Error())
	}
	if newURL, err2 := url.Parse(presignPutObject.URL); err2 == nil {
		return *newURL, nil
	} else {
		return url.URL{}, fmt.Errorf("failed to create presign uploading link with key name %s: parsing URL error: %s",
			fileInfo.FileName, err2.Error())
	}
}

func (s *S3Storage) DownloadFile(fileInfo DownloadFileInfo, expireTime time.Duration) (url.URL, error) {
	presignGetObject, err := s.presignS3.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &fileInfo.FileName,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expireTime
	})
	if err != nil {
		return url.URL{}, fmt.Errorf("failed to create presign downloading link with key name %s: %s",
			fileInfo.FileName, err.Error())
	}
	if newURL, err2 := url.Parse(presignGetObject.URL); err2 == nil {
		return *newURL, nil
	} else {
		return url.URL{}, fmt.Errorf("failed to create presign downloading link with key name %s: %s",
			fileInfo.FileName, err2.Error())
	}
}
