package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/deca/voicepilot-eino/internal/config"
	"github.com/deca/voicepilot-eino/internal/workflow"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var cleanupTicker *time.Ticker

// Handler handles HTTP requests
type Handler struct {
	workflow *workflow.VoiceWorkflow
}

// NewHandler creates a new handler
func NewHandler() *Handler {
	return &Handler{
		workflow: workflow.NewVoiceWorkflow(),
	}
}

// VoiceInteraction handles voice interaction requests
func (h *Handler) VoiceInteraction(c *gin.Context) {
	log.Println("Received voice interaction request")

	// Parse multipart form
	file, err := c.FormFile("audio")
	if err != nil {
		log.Printf("Failed to get audio file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少音频文件",
		})
		return
	}

	// Validate file size
	if file.Size > config.AppConfig.MaxAudioSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("音频文件过大（最大：%d MB）", config.AppConfig.MaxAudioSize/1024/1024),
		})
		return
	}

	// Generate session ID
	sessionID := c.PostForm("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Save audio file to temp directory
	filename := fmt.Sprintf("audio_%d_%s.wav", time.Now().Unix(), sessionID)
	audioPath := filepath.Join(config.AppConfig.TempAudioPath, filename)

	if err := c.SaveUploadedFile(file, audioPath); err != nil {
		log.Printf("Failed to save audio file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存音频文件失败",
		})
		return
	}

	// Clean up temp file after processing
	defer func() {
		if err := os.Remove(audioPath); err != nil {
			log.Printf("Failed to remove temp file: %v", err)
		}
	}()

	// Execute workflow
	response, err := h.workflow.Execute(c.Request.Context(), audioPath, sessionID)
	if err != nil {
		log.Printf("Workflow execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("处理失败：%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// TextInteraction handles text-based interaction requests
func (h *Handler) TextInteraction(c *gin.Context) {
	var req struct {
		Text      string `json:"text" binding:"required"`
		SessionID string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	// Generate session ID if not provided
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	// Execute text-based workflow (skip ASR, start from Intent node)
	response, err := h.workflow.ExecuteText(c.Request.Context(), req.Text, req.SessionID)
	if err != nil {
		log.Printf("Text workflow execution failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("处理失败：%v", err),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

// ServeAudio serves static audio files
func (h *Handler) ServeAudio(c *gin.Context) {
	filename := c.Param("filename")
	filepath := filepath.Join(config.AppConfig.StaticAudioPath, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "音频文件不存在",
		})
		return
	}

	c.File(filepath)
}

// UploadAudio handles audio file upload (for testing)
func (h *Handler) UploadAudio(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少音频文件",
		})
		return
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "打开文件失败",
		})
		return
	}
	defer src.Close()

	// Create destination file
	filename := fmt.Sprintf("upload_%d_%s", time.Now().Unix(), file.Filename)
	dstPath := filepath.Join(config.AppConfig.TempAudioPath, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "创建文件失败",
		})
		return
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "保存文件失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filename": filename,
		"path":     dstPath,
		"size":     file.Size,
	})
}

// StartSessionCleanup starts a background task to periodically clean up expired sessions
func (h *Handler) StartSessionCleanup(interval time.Duration) {
	log.Printf("Starting session cleanup task (interval: %v)", interval)

	cleanupTicker = time.NewTicker(interval)

	go func() {
		for range cleanupTicker.C {
			log.Println("Running session cleanup task...")
			if err := h.workflow.CleanupSessions(); err != nil {
				log.Printf("Session cleanup failed: %v", err)
			} else {
				log.Println("Session cleanup completed")
			}
		}
	}()
}

// StopSessionCleanup stops the background cleanup task
func (h *Handler) StopSessionCleanup() {
	if cleanupTicker != nil {
		cleanupTicker.Stop()
		log.Println("Session cleanup task stopped")
	}
}
