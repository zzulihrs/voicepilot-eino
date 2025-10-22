package qiniu

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/storage"
)

// UploadAudioToStorage uploads audio file to Qiniu Cloud storage and returns public URL
func (c *Client) UploadAudioToStorage(ctx context.Context, localPath string) (string, error) {
	// Get storage credentials from environment
	accessKey := os.Getenv("QINIU_ACCESS_KEY")
	secretKey := os.Getenv("QINIU_SECRET_KEY")
	bucket := os.Getenv("QINIU_BUCKET")
	domain := os.Getenv("QINIU_DOMAIN")

	// Use demo credentials if not configured (for testing with public bucket)
	if accessKey == "" || secretKey == "" {
		log.Println("Storage credentials not configured, ASR will not work")
		return "", fmt.Errorf("需要配置QINIU_ACCESS_KEY和QINIU_SECRET_KEY环境变量才能使用语音识别")
	}

	if bucket == "" {
		bucket = "voicepilot-audio" // default bucket name
	}

	if domain == "" {
		domain = bucket + ".example.com" // user must configure actual domain
	}

	mac := auth.New(accessKey, secretKey)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{
		Zone:          &storage.ZoneHuadong, // 华东区
		UseHTTPS:      true,
		UseCdnDomains: false,
	}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	// Generate unique key for the file
	key := fmt.Sprintf("asr/%d_%s", time.Now().Unix(), filepath.Base(localPath))

	// Upload file
	err := formUploader.PutFile(ctx, &ret, upToken, key, localPath, nil)
	if err != nil {
		log.Printf("Failed to upload file to storage: %v", err)
		return "", fmt.Errorf("上传音频文件失败: %w", err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("https://%s/%s", domain, key)

	log.Printf("Audio uploaded successfully: %s", publicURL)
	return publicURL, nil
}
