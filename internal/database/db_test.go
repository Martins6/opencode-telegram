package database

import (
	"testing"
	"time"
)

func TestGetDueScheduledTasksCutoff(t *testing.T) {
	workspacePath := t.TempDir()
	if err := Init(workspacePath); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	userID := int64(42)

	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	pastID, err := InsertScheduledTask(userID, "in 1m", "echo past", workspacePath, "notify", "notify", &past)
	if err != nil {
		t.Fatalf("failed to insert past task: %v", err)
	}
	defer DeleteScheduledTask(pastID)

	futureID, err := InsertScheduledTask(userID, "in 1h", "echo future", workspacePath, "notify", "notify", &future)
	if err != nil {
		t.Fatalf("failed to insert future task: %v", err)
	}
	defer DeleteScheduledTask(futureID)

	t.Run("cutoff now returns only past", func(t *testing.T) {
		tasks, err := GetDueScheduledTasks(userID, time.Now())
		if err != nil {
			t.Fatalf("GetDueScheduledTasks error: %v", err)
		}
		if len(tasks) != 1 {
			t.Fatalf("got %d tasks, want 1", len(tasks))
		}
		if tasks[0].ID != pastID {
			t.Errorf("got task ID %d, want %d", tasks[0].ID, pastID)
		}
	})

	t.Run("cutoff far future returns both", func(t *testing.T) {
		tasks, err := GetDueScheduledTasks(userID, time.Now().Add(2*time.Hour))
		if err != nil {
			t.Fatalf("GetDueScheduledTasks error: %v", err)
		}
		if len(tasks) != 2 {
			t.Errorf("got %d tasks, want 2", len(tasks))
		}
	})

	t.Run("cutoff far past returns none", func(t *testing.T) {
		tasks, err := GetDueScheduledTasks(userID, time.Now().Add(-2*time.Hour))
		if err != nil {
			t.Fatalf("GetDueScheduledTasks error: %v", err)
		}
		if len(tasks) != 0 {
			t.Errorf("got %d tasks, want 0", len(tasks))
		}
	})
}
