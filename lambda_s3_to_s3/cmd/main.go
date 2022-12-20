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

// customEvent holds an array of S3EventRecord and Bucket/Key.
// Bucket and Key can be used when invoke lambda function by Test event(https://docs.aws.amazon.com/lambda/latest/dg/testing-functions.html).
type customEvent struct {
	S3EventRecords []events.S3EventRecord `json:"Records"`
	Bucket         string                 `json:"bucket,omitempty" example:"shuntagami-demo-data"`
	Key            string                 `json:"key,omitempty" example:"folder1/1.zip"`
}

// handler is Lambda handler func
func handler(ctx context.Context, event customEvent) {
	var bucket, key string

	if len(event.S3EventRecords) != 0 {
		record := event.S3EventRecords[0]
		bucket = record.S3.Bucket.Name
		var err error
		// NOTE: S3 key is URL escaped, so unescaping here
		if key, err = url.QueryUnescape(record.S3.Object.Key); err != nil {
			log.Fatal(err, "E01")
		}
	} else {
		// when lambda func invoked by Test event, use the bucket and key specified by Event JSON(in lambda console)
		bucket, key = event.Bucket, event.Key
	}

	log.Printf("bucket: %s, key: %s", bucket, key)

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

		content, err := ioutil.ReadAll(fc)
		if err != nil {
			log.Fatal(err, "E06")
		}
		func() {
			defer fc.Close()
		}()

		if err := s3Client.UploadWithBytes(ctx, content, bucket, filepath.Join(strings.TrimSuffix(key, filepath.Ext(key)), file.Name), "image/png"); err != nil {
			log.Fatal(err, "E07")
		}
	}
}

// main starts Lambda function
func main() {
	lambda.Start(handler)
}
