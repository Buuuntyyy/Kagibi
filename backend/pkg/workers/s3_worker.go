// Copyright (C) 2025-2026  Buuuntyyy
// SPDX-License-Identifier: AGPL-3.0-or-later

package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"kagibi/backend/pkg/monitoring"
	"kagibi/backend/pkg/s3storage"

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

// Helper for Go < 1.21
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const QueueName = "s3_tasks"

func EnqueueTask(redisClient *redis.Client, task S3Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		log.Printf("EnqueueTask ERROR: Failed to marshal task: %v", err)
		return err
	}
	err = redisClient.RPush(context.Background(), QueueName, data).Err()
	if err != nil {
		log.Printf("EnqueueTask ERROR: Failed to push to Redis: %v", err)
	}
	return err
}

func StartWorker(redisClient *redis.Client) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("S3 Worker PANIC recovered: %v", r)
				// Restart worker on panic
				StartWorker(redisClient)
			}
		}()

		log.Println("S3 Worker started and waiting for tasks...")
		for {
			// Blocking pop from Redis queue
			result, err := redisClient.BLPop(context.Background(), 0, QueueName).Result()
			if err != nil {
				log.Printf("S3 Worker Redis error: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			// Verify result structure
			if len(result) < 2 {
				log.Printf("S3 Worker ERROR: BLPop returned unexpected result: %v", result)
				continue
			}

			payload := result[1]

			// result[0] is the key, result[1] is the value
			var task S3Task
			if err := json.Unmarshal([]byte(payload), &task); err != nil {
				log.Printf("S3 Worker unmarshal error: %v. Payload: %s", err, payload)
				continue
			}

			processTask(task)
		}
	}()
}

func processTask(task S3Task) {
	ctx := context.Background()
	if task.Type == TaskUpload {
		processUploadTask(ctx, task)
		return
	}
	processMoveTask(ctx, task)
}

func processUploadTask(ctx context.Context, task S3Task) {
	if _, err := os.Stat(task.SrcKey); err != nil {
		log.Printf("S3 Worker ERROR: File does not exist! Path: %s, Error: %v", task.SrcKey, err)
		return
	}

	file, err := os.Open(task.SrcKey)
	if err != nil {
		log.Printf("S3 Worker ERROR: Cannot open temp file %s: %v", task.SrcKey, err)
		return
	}
	defer file.Close()

	uploader := manager.NewUploader(s3storage.Client)
	s3Start := time.Now()
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s3storage.BucketName),
		Key:         aws.String(task.DestKey),
		Body:        file,
		ContentType: aws.String(task.ContentType),
	})
	monitoring.RecordS3Duration("put", time.Since(s3Start))
	if err != nil {
		monitoring.RecordS3Request("put", false)
		log.Printf("S3 Worker ERROR: Upload failed for %s: %v", task.DestKey, err)
		return
	}
	monitoring.RecordS3Request("put", true)

	file.Close()
	if err := os.Remove(task.SrcKey); err != nil {
		log.Printf("S3 Worker WARNING: Failed to delete temp file %s: %v", task.SrcKey, err)
	}
}

func processMoveTask(ctx context.Context, task S3Task) {
	if task.IsFolder {
		processFolderMove(ctx, task)
	} else {
		if err := copyAndDelete(ctx, task.SrcKey, task.DestKey); err != nil {
			log.Printf("Error moving file %s: %v", task.SrcKey, err)
		}
	}
}

func processFolderMove(ctx context.Context, task S3Task) {
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
			newKey := strings.Replace(oldKey, srcPrefix, destPrefix, 1)
			if err := copyAndDelete(ctx, oldKey, newKey); err != nil {
				log.Printf("Error moving object %s: %v", oldKey, err)
			}
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
