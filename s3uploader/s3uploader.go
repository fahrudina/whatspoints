package s3uploader

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/wa-serv/config"
)

// UploadToS3 uploads the given data to an S3 bucket and returns the public URL
func UploadToS3(data []byte) (string, error) {
	// Use region and bucket name from the centralized environment configuration
	region := config.Env.AWSRegion
	bucket := config.Env.S3BucketName

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Generate a unique filename
	fileName := uuid.New().String() + ".jpg"

	// Upload the file to S3
	s3Client := s3.New(sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(data), // Use bytes.NewReader to create an io.ReadSeeker
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return the public URL of the uploaded file
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, fileName), nil
}
