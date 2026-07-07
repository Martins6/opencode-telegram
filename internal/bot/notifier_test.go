package bot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/database"
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

func writeConfigFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestConfigHotReload(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	orig := config.Get()
	t.Cleanup(func() { config.SetForTest(orig) })

	writeConfigFile(t, cfgPath, `
[bot]
token = "test-token"
allowed_user_id = "test-user"
timezone = "America/Sao_Paulo"

[defaults]
agent = "agentA"
model = "modelA"
provider = "providerA"

[workspace]
path = "/tmp/test"
`)
	if _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("first load: %v", err)
	}
	if got := config.Get().Defaults.Agent; got != "agentA" {
		t.Fatalf("after first load: agent = %q, want agentA", got)
	}
	if got := config.Get().Bot.Timezone; got != "America/Sao_Paulo" {
		t.Fatalf("after first load: timezone = %q, want America/Sao_Paulo", got)
	}

	writeConfigFile(t, cfgPath, `
[bot]
token = "test-token"
allowed_user_id = "test-user"
timezone = "America/Sao_Paulo"

[defaults]
agent = "agentB"
model = "modelB"
provider = "providerB"

[workspace]
path = "/tmp/test"
`)
	if _, err := config.Load(cfgPath); err != nil {
		t.Fatalf("second load: %v", err)
	}
	if got := config.Get().Defaults.Agent; got != "agentB" {
		t.Errorf("after hot-reload: agent = %q, want agentB (handler would use stale value)", got)
	}
	if got := config.Get().Defaults.Model; got != "modelB" {
		t.Errorf("after hot-reload: model = %q, want modelB", got)
	}
	if got := config.Get().Bot.Timezone; got != "America/Sao_Paulo" {
		t.Errorf("after hot-reload: timezone changed unexpectedly: %q", got)
	}
}

func TestMain(m *testing.M) {
	workspacePath := filepath.Join(os.TempDir(), "test-notifier-db")
	os.MkdirAll(workspacePath, 0755)
	defer os.RemoveAll(workspacePath)

	code := m.Run()
	os.Exit(code)
}
