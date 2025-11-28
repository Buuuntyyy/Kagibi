package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"safercloud/backend/pkg/s3storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
)

type S3TaskType string

const (
	TaskMove   S3TaskType = "move"
	TaskRename S3TaskType = "rename" // Same as move for S3
	TaskUpload S3TaskType = "upload"
)

type S3Task struct {
	Type        S3TaskType `json:"type"`
	UserID      string     `json:"user_id"`
	SrcKey      string     `json:"src_key"`  // For move/rename: old S3 key. For upload: local temp file path.
	DestKey     string     `json:"dest_key"` // For move/rename: new S3 key. For upload: S3 key.
	IsFolder    bool       `json:"is_folder"`
	ContentType string     `json:"content_type"` // For upload
}

const QueueName = "s3_tasks"

func EnqueueTask(redisClient *redis.Client, task S3Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return redisClient.RPush(context.Background(), QueueName, data).Err()
}

func StartWorker(redisClient *redis.Client) {
	go func() {
		log.Println("S3 Worker started")
		for {
			// Blocking pop from Redis queue
			result, err := redisClient.BLPop(context.Background(), 0, QueueName).Result()
			if err != nil {
				log.Printf("Worker Redis error: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			// result[0] is the key, result[1] is the value
			var task S3Task
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				log.Printf("Worker unmarshal error: %v", err)
				continue
			}

			log.Printf("Processing task: %s %s -> %s", task.Type, task.SrcKey, task.DestKey)
			processTask(task)
		}
	}()
}

func processTask(task S3Task) {
	ctx := context.Background()

	if task.Type == TaskUpload {
		// Upload from local temp file to S3
		file, err := os.Open(task.SrcKey)
		if err != nil {
			log.Printf("Error opening temp file for upload %s: %v", task.SrcKey, err)
			return
		}
		defer file.Close()

		uploader := manager.NewUploader(s3storage.Client)
		_, err = uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s3storage.BucketName),
			Key:         aws.String(task.DestKey),
			Body:        file,
			ContentType: aws.String(task.ContentType),
		})

		if err != nil {
			log.Printf("Error uploading file to S3 %s: %v", task.DestKey, err)
			// Retry logic could be added here
			return
		}

		// Delete temp file on success
		file.Close()
		if err := os.Remove(task.SrcKey); err != nil {
			log.Printf("Warning: Failed to delete temp file %s: %v", task.SrcKey, err)
		}
		return
	}

	if task.IsFolder {
		// List objects with prefix SrcKey
		// Note: SrcKey for folder should end with "/"
		srcPrefix := task.SrcKey
		if !strings.HasSuffix(srcPrefix, "/") {
			srcPrefix += "/"
		}
		destPrefix := task.DestKey
		if !strings.HasSuffix(destPrefix, "/") {
			destPrefix += "/"
		}

		paginator := s3.NewListObjectsV2Paginator(s3storage.Client, &s3.ListObjectsV2Input{
			Bucket: aws.String(s3storage.BucketName),
			Prefix: aws.String(srcPrefix),
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				log.Printf("Error listing objects for folder move: %v", err)
				return
			}

			for _, obj := range page.Contents {
				oldKey := *obj.Key
				// Replace prefix
				newKey := strings.Replace(oldKey, srcPrefix, destPrefix, 1)

				err := copyAndDelete(ctx, oldKey, newKey)
				if err != nil {
					log.Printf("Error moving object %s: %v", oldKey, err)
					// Continue with other files? Or stop?
					// For now, continue to try to move as much as possible.
				}
			}
		}

	} else {
		// Single file
		err := copyAndDelete(ctx, task.SrcKey, task.DestKey)
		if err != nil {
			log.Printf("Error moving file %s: %v", task.SrcKey, err)
		}
	}
}

func copyAndDelete(ctx context.Context, srcKey, destKey string) error {
	// 1. Copy
	// CopySource must be URL encoded "bucket/key"
	copySource := fmt.Sprintf("%s/%s", s3storage.BucketName, srcKey)

	_, err := s3storage.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s3storage.BucketName),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	})
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// 2. Delete
	_, err = s3storage.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3storage.BucketName),
		Key:    aws.String(srcKey),
	})
	if err != nil {
		// If delete fails, we have a duplicate. Not critical data loss, but waste of space.
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}
