package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()

	// Load environment variables
	remoteBucket := os.Getenv("REMOTE_BUCKET")
	remoteRegion := os.Getenv("REMOTE_REGION")
	remoteAccessKey := os.Getenv("REMOTE_ACCESS_KEY")
	remoteSecretKey := os.Getenv("REMOTE_SECRET_KEY")
	remoteEndpoint := os.Getenv("REMOTE_ENDPOINT")
	remoteUseSSL := os.Getenv("REMOTE_USE_SSL") != "false"

	localBucket := os.Getenv("LOCAL_BUCKET")
	localRegion := os.Getenv("LOCAL_REGION")
	localAccessKey := os.Getenv("LOCAL_ACCESS_KEY")
	localSecretKey := os.Getenv("LOCAL_SECRET_KEY")
	localEndpoint := os.Getenv("LOCAL_ENDPOINT")
	localUseSSL := os.Getenv("LOCAL_USE_SSL") != "false"

	// Validate required env vars
	if remoteBucket == "" || localBucket == "" || remoteAccessKey == "" || remoteSecretKey == "" ||
		localAccessKey == "" || localSecretKey == "" || localEndpoint == "" || remoteEndpoint == "" {
		log.Fatal("Missing required environment variables")
	}

	// Initialize remote MinIO client
	remoteClient, err := minio.New(remoteEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(remoteAccessKey, remoteSecretKey, ""),
		Region: remoteRegion,
		Secure: remoteUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create remote MinIO client: %v", err)
	}

	// Initialize local MinIO client
	localClient, err := minio.New(localEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(localAccessKey, localSecretKey, ""),
		Region: localRegion,
		Secure: localUseSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create local MinIO client: %v", err)
	}

	// List remote objects
	objectCh := remoteClient.ListObjects(ctx, remoteBucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Failed to list object: %v", object.Err)
			continue
		}

		key := object.Key
		fmt.Printf("Syncing: %s\n", key)

		// Check if object already exists in local bucket
		_, err := localClient.StatObject(ctx, localBucket, key, minio.StatObjectOptions{})
		if err == nil {
			fmt.Printf("Already exists locally: %s\n", key)
			continue
		}

		// Download object from remote
		remoteObj, err := remoteClient.GetObject(ctx, remoteBucket, key, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("Failed to get remote object %s: %v", key, err)
			continue
		}
		defer remoteObj.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, remoteObj)
		if err != nil {
			log.Printf("Failed to read remote object body: %v", err)
			continue
		}

		// Upload to local
		_, err = localClient.PutObject(ctx, localBucket, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), minio.PutObjectOptions{})
		if err != nil {
			log.Printf("Failed to upload to local bucket: %v", err)
			continue
		}

		fmt.Printf("Successfully synced: %s\n", key)
	}
}
