package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinIORepository interface {
	UploadArchive(ctx context.Context, filename string, data []byte) error
	GetArchive(ctx context.Context, filename string) ([]byte, error)
	ListArchives(ctx context.Context, prefix string) ([]string, error)
}

type minIORepository struct {
	client *minio.Client
	bucket string
	log    *logger.Logger
}

func NewMinIORepository(
	endpoint, accessKey, secretKey, bucket string,
	log *logger.Logger,
) (MinIORepository, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info("MinIO bucket created", zap.String("bucket", bucket))
	}

	log.Info("MinIO repository initialized", zap.String("endpoint", endpoint), zap.String("bucket", bucket))

	return &minIORepository{
		client: client,
		bucket: bucket,
		log:    log,
	}, nil
}

func (r *minIORepository) UploadArchive(ctx context.Context, filename string, data []byte) error {
	reader := strings.NewReader(string(data))

	_, err := r.client.PutObject(ctx, r.bucket, filename, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		r.log.Error("Failed to upload archive to MinIO", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("failed to upload archive: %w", err)
	}

	r.log.Info("Archive uploaded to MinIO", zap.String("filename", filename), zap.Int("size", len(data)))
	return nil
}

func (r *minIORepository) GetArchive(ctx context.Context, filename string) ([]byte, error) {
	obj, err := r.client.GetObject(ctx, r.bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		r.log.Error("Failed to get archive from MinIO", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to get archive: %w", err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		r.log.Error("Failed to read archive from MinIO", zap.String("filename", filename), zap.Error(err))
		return nil, fmt.Errorf("failed to read archive: %w", err)
	}

	return data, nil
}

func (r *minIORepository) ListArchives(ctx context.Context, prefix string) ([]string, error) {
	var archives []string

	for obj := range r.client.ListObjects(ctx, r.bucket, minio.ListObjectsOptions{
		Prefix: prefix,
	}) {
		if obj.Err != nil {
			r.log.Error("Error listing objects", zap.Error(obj.Err))
			continue
		}
		archives = append(archives, obj.Key)
	}

	return archives, nil
}

func CreateCSVArchive(menus []*model.MenuItem) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"id", "name", "description", "category", "ingredients", "calories", "proteins", "fats", "carbs", "tags", "created_at", "updated_at"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, menu := range menus {
		row := []string{
			menu.ID.Hex(),
			menu.Name,
			menu.Description,
			menu.Category,
			strings.Join(menu.Ingredients, "|"),
			fmt.Sprintf("%.1f", menu.KBJU.Calories),
			fmt.Sprintf("%.1f", menu.KBJU.Proteins),
			fmt.Sprintf("%.1f", menu.KBJU.Fats),
			fmt.Sprintf("%.1f", menu.KBJU.Carbs),
			strings.Join(menu.Tags, "|"),
			menu.CreatedAt.Format(time.RFC3339),
			menu.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func CreateJSONArchive(menus []*model.MenuItem) ([]byte, error) {
	return json.MarshalIndent(menus, "", "  ")
}
