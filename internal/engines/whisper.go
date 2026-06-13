// Purpose: Whisper transcription via Groq or OpenAI.
// Docs: internal/engines/whisper.doc.md
package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// transcribe calls Whisper via Groq or OpenAI.
func transcribe(ctx context.Context, audioPath, backend string) (string, error) {
	stat, err := os.Stat(audioPath)
	if err != nil {
		return "", err
	}
	if stat.Size() > 25*1024*1024 {
		return "", fmt.Errorf("audio file too large (%.1f MB > 25 MB limit) – use --start/--end", float64(stat.Size())/1024/1024)
	}

	switch backend {
	case "groq":
		return transcribeGroq(ctx, audioPath)
	case "openai":
		return transcribeOpenAI(ctx, audioPath)
	default:
		return "", fmt.Errorf("unsupported whisper backend: %s", backend)
	}
}

func transcribeGroq(ctx context.Context, audioPath string) (string, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}
	return postWhisper(ctx, audioPath, apiKey, "https://api.groq.com/openai/v1/audio/transcriptions", "whisper-large-v3")
}

func transcribeOpenAI(ctx context.Context, audioPath string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}
	return postWhisper(ctx, audioPath, apiKey, "https://api.openai.com/v1/audio/transcriptions", "whisper-1")
}

func postWhisper(ctx context.Context, audioPath, apiKey, url, model string) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	writer.WriteField("model", model)
	writer.WriteField("response_format", "json")
	writer.WriteField("language", "en")
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("whisper %s: %s - %s", model, resp.Status, string(b))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Text, nil
}
