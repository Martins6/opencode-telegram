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

func GetFilePath(workspace string, mediaType MediaType, fileName string) string {
	downloadsDir := filepath.Join(workspace, "downloads", string(mediaType))
	os.MkdirAll(downloadsDir, 0755)
	return filepath.Join(downloadsDir, fileName)
}

func DownloadFile(ctx context.Context, b *bot.Bot, fileID string) ([]byte, error) {
	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.Token(), file.FilePath)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func SaveFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func BuildPrompt(mediaPath string, userMessage string) string {
	return fmt.Sprintf("File located at: %s\n\nUser message: %s", mediaPath, userMessage)
}
