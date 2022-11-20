package infrastructure

import (
	"bytes"
	"context"
	"lambda_s3_to_s3/pkg/helper"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	region = "ap-northeast-1"
)

// AWSS3Operator is interface to manage download/upload for S3
type AWSS3Operator interface {
	// DownloadObject download object at bucket/key on S3 to destPath
	DownloadObject(ctx context.Context, bucket, objectKey, destPath string) (*os.File, error)
	// UploadWithBytes receive []byte and upload it to specified bucket/key
	UploadWithBytes(ctx context.Context, object []byte, bucket, objectKey, contentType string) error
}

type awsS3 struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func (a *awsS3) DownloadObject(ctx context.Context, bucket, objectKey, destPath string) (*os.File, error) {
	if err := helper.EnsureBaseDir(destPath); err != nil {
		return nil, err
	}
	file, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	_, err = a.downloader.DownloadWithContext(ctx, file, &s3.GetObjectInput{
		Bucket: helper.String(bucket),
		Key:    helper.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (a *awsS3) UploadWithBytes(ctx context.Context, object []byte, bucket, objectKey, contentType string) error {
	reader := bytes.NewReader(object)

	_, err := a.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      helper.String(bucket),
		Body:        aws.ReadSeekCloser(reader),
		Key:         helper.String(objectKey),
		ContentType: helper.String(contentType),
	})

	return err
}

// initializeS3Client initialize S3 client
func InitializeS3Client(accessKeyID, secretAccessKey, token string) (AWSS3Operator, error) {
	var cred *credentials.Credentials
	if accessKeyID != "" && secretAccessKey != "" {
		cred = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, token)
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      helper.String(region)},
	)
	if err != nil {
		return nil, err
	}
	return &awsS3{
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}, nil
}
