package main

import (
	"archive/zip"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AWSS3Operator is interface to manage download/upload for S3
type AWSS3Operator interface {
	// downloadObject download object from specified bucket/key
	DownloadObject(bucket string, objectKey string) (*os.File, error)
	// uploadWithBytes receive []byte and upload it to specified bucket/key
	UploadWithBytes(object []byte, bucket string, objectKey string, contentType string) error
}

type awsS3 struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func (a *awsS3) DownloadObject(bucket string, objectKey string) (*os.File, error) {
	// NOTE: /tmp is the only location we are allowed to write to in the Lambda environtment
	zipFile, err := os.Create("/tmp/test.zip")
	if err != nil {
		return nil, err
	}
	_, err = a.downloader.Download(zipFile, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return nil, err
	}
	return zipFile, nil
}

func (a *awsS3) UploadWithBytes(object []byte, bucket string, objectKey string, contentType string) error {
	reader := bytes.NewReader(object)

	_, err := a.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucket),
		Body:        aws.ReadSeekCloser(reader),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return err
	}
	return nil
}

// initializeS3Client initialize S3 client
func initializeS3Client() AWSS3Operator {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1")},
	)
	if err != nil {
		log.Fatal(err)
	}
	return &awsS3{
		uploader:   s3manager.NewUploader(sess),
		downloader: s3manager.NewDownloader(sess),
	}
}

// handler is Lambda handler func
func handler(ctx context.Context, s3Event events.S3Event) {
	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key, err := url.QueryUnescape(record.S3.Object.Key)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := initializeS3Client()
	object, err := s3Client.DownloadObject(bucket, key)
	if err != nil {
		log.Fatal(err)
	}

	zf, err := zip.OpenReader(object.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer zf.Close()

	for _, file := range zf.File {
		if strings.ToLower(filepath.Ext(file.Name)) != ".png" {
			continue
		}

		fc, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer fc.Close()

		content, err := ioutil.ReadAll(fc)
		if err != nil {
			log.Fatal(err)
		}
		if err := s3Client.UploadWithBytes(content, bucket, filepath.Join(strings.TrimSuffix(key, filepath.Ext(key)), file.Name), "image/png"); err != nil {
			log.Fatal(err)
		}
	}
}

// main starts Lambda function
func main() {
	lambda.Start(handler)
}
