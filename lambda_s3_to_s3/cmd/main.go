package main

import (
	"archive/zip"
	"context"
	"io/ioutil"
	"lambda_s3_to_s3/pkg/infrastructure"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// handler is Lambda handler func
func handler(ctx context.Context, s3Event events.S3Event) {
	record := s3Event.Records[0]
	bucket := record.S3.Bucket.Name
	key, err := url.QueryUnescape(record.S3.Object.Key) // S3 key is URL escaped, so unescaping here
	if err != nil {
		log.Fatal(err, "E01")
	}

	s3Client, err := infrastructure.InitializeS3Client("", "", "")
	if err != nil {
		log.Fatal(err, "E02")
	}

	// Download ZIP file
	// NOTE: /tmp is the only location we are allowed to write to in the Lambda environtment
	object, err := s3Client.DownloadObject(ctx, bucket, key, filepath.Join("/tmp", key))
	if err != nil {
		log.Fatal(err, "E03")
	}

	// Open the ZIP file
	zf, err := zip.OpenReader(object.Name())
	if err != nil {
		log.Fatal(err, "E04")
	}
	defer zf.Close()

	// Extract all the files in the ZIP file
	for _, file := range zf.File {
		if strings.ToLower(filepath.Ext(file.Name)) != ".png" {
			continue
		}

		fc, err := file.Open()
		if err != nil {
			log.Fatal(err, "E05")
		}
		defer fc.Close()

		content, err := ioutil.ReadAll(fc)
		if err != nil {
			log.Fatal(err, "E06")
		}

		if err := s3Client.UploadWithBytes(ctx, content, bucket, filepath.Join(strings.TrimSuffix(key, filepath.Ext(key)), file.Name), "image/png"); err != nil {
			log.Fatal(err, "E07")
		}
	}
}

// main starts Lambda function
func main() {
	lambda.Start(handler)
}
