package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/martins6/opencode-telegram/internal/database"
)

func TestNotifyCommandDetection(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	tests := []struct {
		name    string
		command string
		wantMsg string
	}{
		{
			name:    "notify with -m flag",
			command: `opencode-telegram notify -m "Hello World"`,
			wantMsg: "Hello World",
		},
		{
			name:    "notify with quoted message",
			command: `opencode-telegram notify "Hello World"`,
			wantMsg: "Hello World",
		},
		{
			name:    "notify with double quotes",
			command: `opencode-telegram notify "Test message"`,
			wantMsg: "Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextRun := time.Now().Add(time.Minute)
			taskID, err := database.InsertScheduledTask(
				userID,
				"in 1m",
				tt.command,
				workspacePath,
				"notify",
				"notify",
				&nextRun,
			)
			if err != nil {
				t.Fatalf("failed to insert task: %v", err)
			}

			task, err := database.GetScheduledTask(taskID)
			if err != nil {
				t.Fatalf("failed to get task: %v", err)
			}

			if task.Command != tt.command {
				t.Errorf("got command %q, want %q", task.Command, tt.command)
			}

			database.DeleteScheduledTask(taskID)
		})
	}
}

func TestMailCommandDetection(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	nextRun := time.Now().Add(time.Hour)
	taskID, err := database.InsertScheduledTask(
		userID,
		"0 9 * * *",
		"opencode-telegram mail send --sender test@example.com --subject Test --content Hello",
		workspacePath,
		"mail",
		"notify",
		&nextRun,
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	task, err := database.GetScheduledTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	if task.Command == "" {
		t.Error("expected command to be stored")
	}

	database.DeleteScheduledTask(taskID)
}

func TestScheduleParsing(t *testing.T) {
	tests := []struct {
		expr    string
		wantErr bool
	}{
		{"in 30m", false},
		{"in 1h", false},
		{"at 09:00", false},
		{"once 14:30", false},
		{"0 9 * * *", false},
		{"*/15 * * * *", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			workspacePath := t.TempDir()
			if err := database.Init(workspacePath); err != nil {
				t.Fatalf("failed to init database: %v", err)
			}

			nextRun := time.Now().Add(time.Hour)
			_, err := database.InsertScheduledTask(
				1,
				tt.expr,
				"echo hello",
				workspacePath,
				"notify",
				"notify",
				&nextRun,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSchedule(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
		})
	}
}

func TestSchedulerTaskExecution(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	nextRun := time.Now().Add(-time.Minute)
	taskID, err := database.InsertScheduledTask(
		userID,
		"in 1m",
		"opencode-telegram notify -m \"Test notification\"",
		workspacePath,
		"notify",
		"notify",
		&nextRun,
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	tasks, err := database.GetDueScheduledTasks(userID)
	if err != nil {
		t.Fatalf("failed to get due tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("expected at least one due task")
	}

	found := false
	for _, task := range tasks {
		if task.ID == taskID {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected task to be due")
	}

	notifications, err := database.GetUnsentNotifications(userID)
	if err != nil {
		t.Fatalf("failed to get notifications: %v", err)
	}

	initialCount := len(notifications)

	task := tasks[0]
	isOneTime := task.ScheduleExpr == "in 1m" || task.ScheduleExpr == "once"

	if isOneTime {
		database.UpdateScheduledTaskStatus(task.ID, "completed")
	}

	notifications, err = database.GetUnsentNotifications(userID)
	if err != nil {
		t.Fatalf("failed to get notifications: %v", err)
	}

	if len(notifications) != initialCount {
		t.Logf("Notification was processed (current count: %d, initial: %d)", len(notifications), initialCount)
	}
}

func TestWorkingDirectory(t *testing.T) {
	testDir := t.TempDir()

	workspacePath := t.TempDir()
	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	nextRun := time.Now().Add(time.Hour)
	_, err := database.InsertScheduledTask(
		1,
		"0 9 * * *",
		"pwd",
		testDir,
		"notify",
		"notify",
		&nextRun,
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	tasks, err := database.ListScheduledTasks(1)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) == 0 {
		t.Fatal("expected at least one task")
	}

	if tasks[0].WorkingDir != testDir {
		t.Errorf("got working dir %q, want %q", tasks[0].WorkingDir, testDir)
	}
}

func TestMailContentParsing(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	nextRun := time.Now().Add(time.Hour)
	mailContent := "Hello World with spaces"
	taskID, err := database.InsertScheduledTask(
		userID,
		"0 9 * * *",
		`opencode-telegram mail send --sender test@example.com --subject Test --content "`+mailContent+`"`,
		workspacePath,
		"mail",
		"notify",
		&nextRun,
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}
	defer database.DeleteScheduledTask(taskID)

	task, err := database.GetScheduledTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	args := parseMailCommandArgs(task.Command)
	if args["--content"] != mailContent {
		t.Errorf("got content %q, want %q", args["--content"], mailContent)
	}
}

func TestMailContentWithEmoji(t *testing.T) {
	workspacePath := t.TempDir()

	if err := database.Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(1)

	nextRun := time.Now().Add(time.Hour)
	mailContent := "📬 Success! Your mail notification arrived right on time!"
	taskID, err := database.InsertScheduledTask(
		userID,
		"0 9 * * *",
		`opencode-telegram mail send --sender Sunny --subject "Test Mail" --content "`+mailContent+`"`,
		workspacePath,
		"mail",
		"notify",
		&nextRun,
	)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}
	defer database.DeleteScheduledTask(taskID)

	task, err := database.GetScheduledTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}

	args := parseMailCommandArgs(task.Command)
	if args["--content"] != mailContent {
		t.Errorf("got content %q, want %q", args["--content"], mailContent)
	}
	if args["--sender"] != "Sunny" {
		t.Errorf("got sender %q, want %q", args["--sender"], "Sunny")
	}
}

func TestMain(m *testing.M) {
	workspacePath := filepath.Join(os.TempDir(), "test-scheduler-db")
	os.MkdirAll(workspacePath, 0755)
	defer os.RemoveAll(workspacePath)

	code := m.Run()
	os.Exit(code)
}
