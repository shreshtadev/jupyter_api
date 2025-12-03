package uploader

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"shreshtasmg.in/jupyter/internal/company"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Service interface {
	GeneratePresignedUploadURL(
		ctx context.Context,
		company *company.Company,
		objectKey string,
		fileSize int64,
	) (string, error)

	DeleteObject(
		ctx context.Context,
		company *company.Company,
		objectKey string,
	) error

	DeletePrefix(
		ctx context.Context,
		company *company.Company,
		prefix string,
	) (deletedCount int, deletedBytes int64, err error)

	ListPrefixes(ctx context.Context, companyRec *company.Company, fullPrefix string, limit int, nextToken string) ([]string, *string, error)
	ListFilesInFolder(ctx context.Context, companyRec *company.Company, folderPrefix string, limit int, nextToken string) ([]string, *string, error)
}

type s3Service struct{}

func NewS3Service() S3Service {
	return &s3Service{}
}

func buildS3Client(ctx context.Context, companyRec *company.Company) (*s3.Client, error) {
	if companyRec.AwsBucketName == nil ||
		companyRec.AwsBucketRegion == nil ||
		companyRec.AwsAccessKey == nil ||
		companyRec.AwsSecretKey == nil {
		return nil, fmt.Errorf("company AWS configuration is incomplete")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(*companyRec.AwsBucketRegion),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				*companyRec.AwsAccessKey,
				*companyRec.AwsSecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return s3.NewFromConfig(awsCfg), nil
}

func (s *s3Service) GeneratePresignedUploadURL(
	ctx context.Context,
	companyRec *company.Company,
	objectKey string,
	fileSize int64,
) (string, error) {
	s3Client, err := buildS3Client(ctx, companyRec)
	if err != nil {
		return "", err
	}

	presigner := s3.NewPresignClient(s3Client)
	out, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(*companyRec.AwsBucketName),
		Key:           aws.String(objectKey),
		ContentLength: aws.Int64(fileSize),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return out.URL, nil
}

func (s *s3Service) DeleteObject(
	ctx context.Context,
	companyRec *company.Company,
	objectKey string,
) error {
	client, err := buildS3Client(ctx, companyRec)
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(*companyRec.AwsBucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// DeletePrefix deletes all objects under the given prefix.
// NOTE: this is a simple, single-page implementation. For very large folders,
// you should paginate ListObjectsV2.
func (s *s3Service) DeletePrefix(
	ctx context.Context,
	companyRec *company.Company,
	prefix string,
) (int, int64, error) {
	client, err := buildS3Client(ctx, companyRec)
	if err != nil {
		return 0, 0, err
	}

	listOut, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(*companyRec.AwsBucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to list objects: %w", err)
	}

	if len(listOut.Contents) == 0 {
		return 0, 0, nil
	}

	mandatory_delete := len(listOut.Contents) >= 10

	objects := make([]types.ObjectIdentifier, 0, len(listOut.Contents))
	var totalBytes int64

	for _, obj := range listOut.Contents {
		objects = append(objects, types.ObjectIdentifier{
			Key: obj.Key,
		})
		if obj.Size != nil {
			totalBytes += int64(*obj.Size)
		}
	}
	if mandatory_delete {
		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(*companyRec.AwsBucketName),
			Delete: &types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return 0, 0, fmt.Errorf("failed to delete objects: %w", err)
		}
	}
	return len(objects), totalBytes, nil
}

func (s *s3Service) ListPrefixes(ctx context.Context, companyRec *company.Company, fullPrefix string, limit int, nextToken string) ([]string, *string, error) {
	s3Client, err := buildS3Client(ctx, companyRec)
	if err != nil {
		return nil, nil, err
	}
	input := &s3.ListObjectsV2Input{
		Bucket:    companyRec.AwsBucketName,
		Prefix:    aws.String(fullPrefix),
		Delimiter: aws.String("/"), // Assuming you applied the fix from the previous advice
		MaxKeys:   aws.Int32(int32(limit)),
	}

	// Only set ContinuationToken if the input token is provided and non-empty
	if nextToken != "" {
		input.ContinuationToken = aws.String(nextToken)
	}

	listOut, err := s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		fmt.Printf("DEBUG S3 Error: %v\n", err)
		return nil, nil, fmt.Errorf("failed to list prefixes: %w", err)
	}

	var prefixes []string
	for _, obj := range listOut.Contents {
		if obj.Key != nil {
			prefix := filepath.Dir(*obj.Key)
			if !contains(prefixes, prefix) {
				prefixes = append(prefixes, prefix)
			}
		}
	}

	return prefixes, listOut.NextContinuationToken, nil
}

func (s *s3Service) ListFilesInFolder(ctx context.Context, companyRec *company.Company, folderPrefix string, limit int, nextToken string) ([]string, *string, error) {
	client, err := buildS3Client(ctx, companyRec)
	if err != nil {
		return nil, nil, err
	}

	// Ensure prefix ends with '/' so we list objects inside the folder
	if folderPrefix != "" && folderPrefix[len(folderPrefix)-1] != '/' {
		folderPrefix = folderPrefix + "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket:    companyRec.AwsBucketName,
		Prefix:    aws.String(folderPrefix),
		Delimiter: aws.String("/"), // ensures subfolders appear in CommonPrefixes, contents are files directly under prefix
		MaxKeys:   aws.Int32(int32(limit)),
	}

	if nextToken != "" {
		input.ContinuationToken = aws.String(nextToken)
	}

	out, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list files: %w", err)
	}

	var files []string
	for _, obj := range out.Contents {
		if obj.Key == nil {
			continue
		}
		k := *obj.Key
		// skip folder placeholder objects and the folder key itself
		if k == folderPrefix {
			continue
		}
		if len(k) > 0 && k[len(k)-1] == '/' {
			continue
		}
		files = append(files, k)
	}

	return files, out.NextContinuationToken, nil
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
