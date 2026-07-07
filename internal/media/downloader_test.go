package media

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/go-telegram/bot/models"
)

func TestGetMediaType_Photo(t *testing.T) {
	msg := &models.Message{
		Photo: []models.PhotoSize{{FileID: "abc", Width: 100, Height: 100}},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypePhoto {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypePhoto)
	}
	if ext != ".jpg" {
		t.Errorf("ext = %q, want %q", ext, ".jpg")
	}
}

func TestGetMediaType_Audio(t *testing.T) {
	msg := &models.Message{
		Audio: &models.Audio{FileID: "abc"},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypeAudio {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypeAudio)
	}
	if ext != ".mp3" {
		t.Errorf("ext = %q, want %q", ext, ".mp3")
	}
}

func TestGetMediaType_Voice(t *testing.T) {
	msg := &models.Message{
		Voice: &models.Voice{FileID: "abc"},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypeVoice {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypeVoice)
	}
	if ext != ".ogg" {
		t.Errorf("ext = %q, want %q", ext, ".ogg")
	}
}

func TestGetMediaType_Document(t *testing.T) {
	msg := &models.Message{
		Document: &models.Document{FileID: "abc", FileName: "report.pdf"},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypeDocument {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypeDocument)
	}
	if ext != ".pdf" {
		t.Errorf("ext = %q, want %q", ext, ".pdf")
	}
}

func TestGetMediaType_DocumentNoName(t *testing.T) {
	msg := &models.Message{
		Document: &models.Document{FileID: "abc"},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypeDocument {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypeDocument)
	}
	if ext != ".bin" {
		t.Errorf("ext = %q, want %q", ext, ".bin")
	}
}

func TestGetMediaType_Video(t *testing.T) {
	msg := &models.Message{
		Video: &models.Video{FileID: "abc"},
	}
	mt, ext, err := GetMediaType(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mt != MediaTypeVideo {
		t.Errorf("MediaType = %q, want %q", mt, MediaTypeVideo)
	}
	if ext != ".mp4" {
		t.Errorf("ext = %q, want %q", ext, ".mp4")
	}
}

func TestGetMediaType_Empty(t *testing.T) {
	msg := &models.Message{}
	_, _, err := GetMediaType(msg)
	if err == nil {
		t.Fatal("expected error for empty message, got nil")
	}
}

func TestGetFilePath(t *testing.T) {
	ws := t.TempDir()
	path := GetFilePath(ws, MediaTypePhoto, ".jpg")
	dir := filepath.Join(ws, "downloads", "images")
	if !strings.HasPrefix(path, dir+string(filepath.Separator)) {
		t.Errorf("GetFilePath = %q, want prefix %q", path, dir)
	}
	if !strings.HasSuffix(path, ".jpg") {
		t.Errorf("GetFilePath = %q, want suffix .jpg", path)
	}
	re := regexp.MustCompile(`\d{8}_\d{6}\.jpg$`)
	if !re.MatchString(path) {
		t.Errorf("GetFilePath = %q, want to match timestamp pattern", path)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("downloads/images directory was not created: %v", err)
	}
}

func TestGenerateTimestampedFilename(t *testing.T) {
	dir := t.TempDir()
	re := regexp.MustCompile(`^\d{8}_\d{6}\.jpg$`)

	t.Run("basic", func(t *testing.T) {
		name := GenerateTimestampedFilename(dir, ".jpg")
		if !re.MatchString(name) {
			t.Errorf("GenerateTimestampedFilename = %q, want to match %s", name, re)
		}
	})

	t.Run("collision", func(t *testing.T) {
		name1 := GenerateTimestampedFilename(dir, ".jpg")
		if err := os.WriteFile(filepath.Join(dir, name1), []byte("x"), 0644); err != nil {
			t.Fatalf("failed to seed collision file: %v", err)
		}
		name2 := GenerateTimestampedFilename(dir, ".jpg")
		ext := ".jpg"
		expected := name1[:len(name1)-len(ext)] + "_1" + ext
		if name2 != expected {
			t.Errorf("GenerateTimestampedFilename on collision = %q, want %q", name2, expected)
		}
		if _, err := os.Stat(filepath.Join(dir, name2)); err == nil {
			t.Errorf("expected %s not to exist yet", name2)
		}
	})
}

func TestBuildPrompt(t *testing.T) {
	got := BuildPrompt("/tmp/x.jpg", "hi")
	want := "File located at: /tmp/x.jpg\n\nUser message: hi"
	if got != want {
		t.Errorf("BuildPrompt = %q, want %q", got, want)
	}
}

func TestBuildPrompt_EmptyCaption(t *testing.T) {
	got := BuildPrompt("/tmp/x.jpg", "")
	want := "File located at: /tmp/x.jpg\n\nUser message: "
	if got != want {
		t.Errorf("BuildPrompt = %q, want %q", got, want)
	}
}
