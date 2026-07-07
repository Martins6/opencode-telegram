package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MediaType string

const (
	MediaTypePhoto    MediaType = "images"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeVoice    MediaType = "audio"
	MediaTypeDocument MediaType = "documents"
	MediaTypeVideo    MediaType = "videos"
)

func GetMediaType(message *models.Message) (MediaType, string, error) {
	switch {
	case message.Photo != nil:
		return MediaTypePhoto, ".jpg", nil
	case message.Audio != nil:
		return MediaTypeAudio, ".mp3", nil
	case message.Voice != nil:
		return MediaTypeVoice, ".ogg", nil
	case message.Document != nil:
		ext := ".bin"
		if message.Document.FileName != "" {
			ext = filepath.Ext(message.Document.FileName)
		}
		return MediaTypeDocument, ext, nil
	case message.Video != nil:
		return MediaTypeVideo, ".mp4", nil
	default:
		return "", "", fmt.Errorf("unknown media type")
	}
}

func GetFilePath(workspace string, mediaType MediaType, ext string) string {
	downloadsDir := filepath.Join(workspace, "downloads", string(mediaType))
	os.MkdirAll(downloadsDir, 0755)
	return filepath.Join(downloadsDir, GenerateTimestampedFilename(downloadsDir, ext))
}

func DownloadFile(ctx context.Context, b *bot.Bot, fileID string) ([]byte, string, error) {
	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), file.FilePath)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file body: %w", err)
	}

	return data, file.FilePath, nil
}

func SaveFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func BuildPrompt(mediaPath string, userMessage string) string {
	return fmt.Sprintf("File located at: %s\n\nUser message: %s", mediaPath, userMessage)
}

func GenerateTimestampedFilename(dir, ext string) string {
	base := time.Now().UTC().Format("20060102_150405")
	name := base + ext
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return name
		}
		name = fmt.Sprintf("%s_%d%s", base, i, ext)
	}
}
