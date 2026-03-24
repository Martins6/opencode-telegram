package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/martins6/opencode-telegram/internal/database"
)

func TestNotifierNotificationDelivery(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	msgID, err := database.InsertNotification(userID, "Test notification")
	if err != nil {
		t.Fatalf("failed to insert notification: %v", err)
	}

	notifications, err := database.GetUnsentNotifications(userID)
	if err != nil {
		t.Fatalf("failed to get notifications: %v", err)
	}

	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifications))
	}

	if notifications[0].Message != "Test notification" {
		t.Errorf("got message %q, want %q", notifications[0].Message, "Test notification")
	}

	if err := database.MarkNotificationSent(msgID); err != nil {
		t.Fatalf("failed to mark notification sent: %v", err)
	}

	notifications, err = database.GetUnsentNotifications(userID)
	if err != nil {
		t.Fatalf("failed to get notifications: %v", err)
	}

	if len(notifications) != 0 {
		t.Errorf("expected 0 unsent notifications after marking sent, got %d", len(notifications))
	}
}

func TestMockSender(t *testing.T) {
	sender := &mockSender{}

	result, err := sender.SendMessage(nil, nil)
	if err != nil {
		t.Errorf("mock sender returned error: %v", err)
	}
	if result != nil {
		t.Error("mock sender should return nil")
	}
}

func TestMockSenderWithParams(t *testing.T) {
	sender := &mockSender{}

	result, err := sender.SendMessage(
		nil,
		nil,
	)
	if err != nil {
		t.Errorf("mock sender returned error: %v", err)
	}
	if result != nil {
		t.Error("mock sender should return nil")
	}
}

func generateTestUUID() string {
	return "test-" + time.Now().Format("20060102150405")
}

func TestMain(m *testing.M) {
	workspacePath := filepath.Join(os.TempDir(), "test-notifier-db")
	os.MkdirAll(workspacePath, 0755)
	defer os.RemoveAll(workspacePath)

	code := m.Run()
	os.Exit(code)
}
