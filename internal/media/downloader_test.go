package media

import (
	"os"
	"path/filepath"
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
	path := GetFilePath(ws, MediaTypePhoto, "x.jpg")
	want := filepath.Join(ws, "downloads", "images", "x.jpg")
	if path != want {
		t.Errorf("GetFilePath = %q, want %q", path, want)
	}
	if _, err := os.Stat(filepath.Join(ws, "downloads", "images")); err != nil {
		t.Errorf("downloads/images directory was not created: %v", err)
	}
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

func TestFilenameFromTelegramPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"photos/file_42.jpg", "file_42.jpg"},
		{"documents/report.pdf", "report.pdf"},
		{"file_42.jpg", "file_42.jpg"},
		{"a/b/c/d.txt", "d.txt"},
		{"", "."},
		{"/", "/"},
	}
	for _, tc := range cases {
		got := FilenameFromTelegramPath(tc.in)
		if got != tc.want {
			t.Errorf("FilenameFromTelegramPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}