package qiniu

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// WebSocket ASR endpoint
	wsASRURL = "wss://openai.qiniu.com/v1/voice/asr"

	// Protocol constants
	protocolVersion = 0x1 // Version 1
	headerSize      = 0x1 // 1 word (4 bytes)

	// Message types
	msgTypeFullClientRequest  = 0x1 // 0b0001 - Full client request
	msgTypeAudioOnlyRequest   = 0x2 // 0b0010 - Audio-only data
	msgTypeFullServiceResponse = 0x9 // 0b1001 - Full service response

	// Serialization methods
	serializationJSON = 0x1 // JSON serialization

	// Compression methods
	compressionNone = 0x0 // No compression
	compressionGzip = 0x1 // Gzip compression

	// Audio parameters
	audioChunkSize = 3200 // 0.2 seconds at 16kHz 16-bit mono (16000 * 2 * 0.2)
)

// WSASRConfig represents the initial configuration for WebSocket ASR
type WSASRConfig struct {
	User    WSASRUser    `json:"user"`
	Audio   WSASRAudio   `json:"audio"`
	Request WSASRRequest `json:"request"`
}

type WSASRUser struct {
	UID string `json:"uid"`
}

type WSASRAudio struct {
	Format     string `json:"format"`      // "pcm"
	SampleRate int    `json:"sample_rate"` // 16000
	Bits       int    `json:"bits"`        // 16
	Channel    int    `json:"channel"`     // 1
	Codec      string `json:"codec"`       // "raw"
}

type WSASRRequest struct {
	ModelName  string `json:"model_name"`  // "asr"
	EnablePunc bool   `json:"enable_punc"` // true
}

// WSASRResponse represents the WebSocket ASR response
type WSASRResponse struct {
	Code      int              `json:"code"`
	Message   string           `json:"message"`
	Reqid     string           `json:"reqid"`
	Result    WSASRResult      `json:"result"`
	AudioInfo WSASRAudioInfo   `json:"audio_info"`
}

type WSASRResult struct {
	Text string `json:"text"`
}

type WSASRAudioInfo struct {
	Duration int `json:"duration"`
}

// buildFrame constructs a binary frame according to the protocol
func buildFrame(msgType byte, serializationMethod byte, compressionMethod byte, sequence int32, payload []byte) []byte {
	// Build 4-byte header
	header := make([]byte, 4)
	header[0] = (protocolVersion << 4) | headerSize               // Protocol version | Header size
	header[1] = (msgType << 4) | 0x0                              // Message type | Flags
	header[2] = (serializationMethod << 4) | compressionMethod    // Serialization | Compression
	header[3] = 0x0                                               // Reserved

	// Build frame: header + sequence + payload_size + payload
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, header)
	binary.Write(buf, binary.BigEndian, sequence)
	binary.Write(buf, binary.BigEndian, int32(len(payload)))
	buf.Write(payload)

	return buf.Bytes()
}

// compressPayload compresses data using gzip
func compressPayload(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// decompressPayload decompresses gzip data
func decompressPayload(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// parseFrame parses a binary frame and returns message type, sequence, and payload
func parseFrame(frame []byte) (msgType byte, sequence int32, payload []byte, err error) {
	if len(frame) < 12 {
		return 0, 0, nil, fmt.Errorf("frame too short: %d bytes", len(frame))
	}

	// Parse header
	msgType = (frame[1] >> 4) & 0x0F
	compressionMethod := frame[2] & 0x0F

	// Parse sequence and payload size
	buf := bytes.NewReader(frame[4:])
	binary.Read(buf, binary.BigEndian, &sequence)
	var payloadSize int32
	binary.Read(buf, binary.BigEndian, &payloadSize)

	// Extract payload
	payload = frame[12:]

	// Decompress if needed
	if compressionMethod == compressionGzip {
		payload, err = decompressPayload(payload)
		if err != nil {
			return 0, 0, nil, fmt.Errorf("failed to decompress payload: %w", err)
		}
	}

	return msgType, sequence, payload, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// WebSocketASR performs speech-to-text conversion using WebSocket
func (c *Client) WebSocketASR(ctx context.Context, audioPath string) (string, error) {
	log.Printf("Starting WebSocket ASR for audio file: %s", audioPath)

	// Read audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	// Detect and convert audio format to PCM if needed
	audioFormat := detectAudioFormat(audioData)
	log.Printf("Detected audio format: %s", audioFormat)

	var pcmData []byte
	if audioFormat != "wav" {
		log.Printf("Converting %s to WAV format...", audioFormat)
		convertedPath, err := convertToWav(audioPath)
		if err != nil {
			return "", fmt.Errorf("failed to convert audio: %w", err)
		}
		defer os.Remove(convertedPath)

		// Read converted WAV file
		audioData, err = os.ReadFile(convertedPath)
		if err != nil {
			return "", fmt.Errorf("failed to read converted audio: %w", err)
		}
	}

	// Extract PCM data from WAV (skip 44-byte WAV header)
	if len(audioData) > 44 && string(audioData[0:4]) == "RIFF" {
		pcmData = audioData[44:]
	} else {
		pcmData = audioData
	}

	log.Printf("PCM data size: %d bytes", len(pcmData))

	// Establish WebSocket connection
	header := http.Header{}
	header.Add("Authorization", "Bearer "+c.apiKey)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsASRURL, header)
	if err != nil {
		return "", fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	log.Println("WebSocket connection established")

	// Prepare configuration
	config := WSASRConfig{
		User: WSASRUser{
			UID: uuid.New().String(),
		},
		Audio: WSASRAudio{
			Format:     "pcm",
			SampleRate: 16000,
			Bits:       16,
			Channel:    1,
			Codec:      "raw",
		},
		Request: WSASRRequest{
			ModelName:  "asr",
			EnablePunc: true,
		},
	}

	// Serialize configuration
	configJSON, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	log.Printf("Sending configuration: %s", string(configJSON))

	// Compress configuration
	compressedConfig, err := compressPayload(configJSON)
	if err != nil {
		return "", fmt.Errorf("failed to compress config: %w", err)
	}

	// Build and send configuration frame
	configFrame := buildFrame(msgTypeFullClientRequest, serializationJSON, compressionGzip, 0, compressedConfig)
	err = conn.WriteMessage(websocket.BinaryMessage, configFrame)
	if err != nil {
		return "", fmt.Errorf("failed to send config frame: %w", err)
	}

	log.Println("Configuration frame sent")

	// Channel to collect results
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start goroutine to read responses
	go func() {
		var fullText string
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Println("WebSocket closed normally")
					resultChan <- fullText
					return
				}
				errChan <- fmt.Errorf("failed to read message: %w", err)
				return
			}

			// Parse frame
			msgType, sequence, payload, err := parseFrame(message)
			if err != nil {
				log.Printf("Failed to parse frame: %v", err)
				continue
			}

			log.Printf("Received message type: 0x%x, sequence: %d, payload size: %d bytes", msgType, sequence, len(payload))
			log.Printf("Payload content: %s", string(payload)[:min(200, len(payload))])

			// Parse response
			if msgType == msgTypeFullServiceResponse {
				var response WSASRResponse
				err = json.Unmarshal(payload, &response)
				if err != nil {
					log.Printf("Failed to unmarshal response: %v", err)
					continue
				}

				log.Printf("ASR response: %+v", response)

				if response.Result.Text != "" {
					fullText = response.Result.Text
					log.Printf("Recognized text: %s", fullText)
				}

				// Check if this is the final response (negative sequence indicates end)
				if sequence < 0 {
					resultChan <- fullText
					return
				}
			}
		}
	}()

	// Send audio data in chunks
	sequence := int32(1)
	for offset := 0; offset < len(pcmData); offset += audioChunkSize {
		end := offset + audioChunkSize
		if end > len(pcmData) {
			end = len(pcmData)
		}

		chunk := pcmData[offset:end]
		log.Printf("Sending audio chunk %d: %d bytes", sequence, len(chunk))

		// Build and send audio frame (no compression for audio data)
		audioFrame := buildFrame(msgTypeAudioOnlyRequest, serializationJSON, compressionNone, sequence, chunk)
		err = conn.WriteMessage(websocket.BinaryMessage, audioFrame)
		if err != nil {
			return "", fmt.Errorf("failed to send audio frame: %w", err)
		}

		sequence++

		// Small delay to avoid overwhelming the server
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("All audio chunks sent")

	// Send close message to indicate end of audio
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// Wait for result or error
	select {
	case result := <-resultChan:
		if result == "" {
			return "", fmt.Errorf("no recognition result received")
		}
		return result, nil
	case err := <-errChan:
		return "", err
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("timeout waiting for ASR result")
	}
}
