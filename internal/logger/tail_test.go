package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeLog(t *testing.T, ws, dateStr, content string) {
	t.Helper()
	dir := filepath.Join(ws, ".logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, dateStr+".log"), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", dateStr, err)
	}
}

func TestTailLastN_MultiDay_ChronologicalOrder(t *testing.T) {
	ws := t.TempDir()

	day1 := strings.Join([]string{
		`[INPUT]  [2026-07-19 09:00:00] User 123: hello-d1`,
		`[OUTPUT]  [2026-07-19 09:00:05] User 123: world-d1`,
		`continuation of world-d1`,
		`[INPUT]  [2026-07-19 09:00:10] User 123: bye-d1`,
		``,
	}, "\n")
	writeLog(t, ws, "2026-07-19", day1)

	day2 := strings.Join([]string{
		`[INPUT]  [2026-07-20 09:00:00] User 123: day2-1`,
		`[INPUT]  [2026-07-20 09:00:05] User 123: day2-2`,
		`[OUTPUT]  [2026-07-20 09:00:10] User 123: day2-3`,
		`[DEBUG]  [2026-07-20 09:00:15] multi`,
		`line debug`,
		`with multiple lines`,
		`[INPUT]  [2026-07-20 09:00:20] User 123: day2-5`,
		``,
	}, "\n")
	writeLog(t, ws, "2026-07-20", day2)

	if err := os.WriteFile(filepath.Join(ws, ".logs", "launchd.out.log"), []byte("ignore me\n"), 0o644); err != nil {
		t.Fatalf("write launchd: %v", err)
	}

	entries, err := TailLastN(ws, 10, time.Time{})
	if err != nil {
		t.Fatalf("TailLastN: %v", err)
	}
	if len(entries) != 8 {
		t.Fatalf("got %d entries, want 8", len(entries))
	}

	var stamps []string
	for _, e := range entries {
		stamps = append(stamps, e.Time.Format("2006-01-02 15:04:05"))
	}
	wantStamps := []string{
		"2026-07-19 09:00:00",
		"2026-07-19 09:00:05",
		"2026-07-19 09:00:10",
		"2026-07-20 09:00:00",
		"2026-07-20 09:00:05",
		"2026-07-20 09:00:10",
		"2026-07-20 09:00:15",
		"2026-07-20 09:00:20",
	}
	for i, w := range wantStamps {
		if stamps[i] != w {
			t.Errorf("entry %d timestamp = %s, want %s", i, stamps[i], w)
		}
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].Time.Before(entries[i-1].Time) {
			t.Errorf("entries out of chronological order at index %d: %v then %v", i, entries[i-1].Time, entries[i].Time)
		}
	}

	if !strings.Contains(entries[1].Raw, "continuation of world-d1") {
		t.Errorf("continuation lines should belong to previous entry Raw, got %q", entries[1].Raw)
	}
	if !strings.Contains(entries[6].Raw, "with multiple lines") {
		t.Errorf("multi-line debug Raw should contain continuation lines, got %q", entries[6].Raw)
	}
}

func TestTailLastN_HonorsNAcrossFiles(t *testing.T) {
	ws := t.TempDir()

	day1 := strings.Join([]string{
		`[INPUT]  [2026-07-19 09:00:00] User 123: d1-a`,
		`[INPUT]  [2026-07-19 09:00:01] User 123: d1-b`,
		`[INPUT]  [2026-07-19 09:00:02] User 123: d1-c`,
		``,
	}, "\n")
	writeLog(t, ws, "2026-07-19", day1)

	day2 := strings.Join([]string{
		`[INPUT]  [2026-07-20 09:00:00] User 123: d2-a`,
		`[INPUT]  [2026-07-20 09:00:01] User 123: d2-b`,
		`[INPUT]  [2026-07-20 09:00:02] User 123: d2-c`,
		`[INPUT]  [2026-07-20 09:00:03] User 123: d2-d`,
		`[INPUT]  [2026-07-20 09:00:04] User 123: d2-e`,
		``,
	}, "\n")
	writeLog(t, ws, "2026-07-20", day2)

	entries, err := TailLastN(ws, 3, time.Time{})
	if err != nil {
		t.Fatalf("TailLastN: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	wantMessages := []string{"d2-c", "d2-d", "d2-e"}
	for i, want := range wantMessages {
		if !strings.Contains(entries[i].Message, want) {
			t.Errorf("entry %d message = %q, want to contain %q", i, entries[i].Message, want)
		}
	}
}

func TestTailLastN_EmptyWorkspace(t *testing.T) {
	ws := t.TempDir()
	entries, err := TailLastN(ws, 10, time.Time{})
	if err != nil {
		t.Fatalf("TailLastN: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("got %d entries, want 0", len(entries))
	}
}

func TestParseDay_MissingFile(t *testing.T) {
	ws := t.TempDir()
	date := time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)
	_, err := ParseDay(ws, date)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "2026-07-20") {
		t.Errorf("error should mention the date 2026-07-20, got %q", err.Error())
	}
}

func TestParseDay_SingleDay(t *testing.T) {
	ws := t.TempDir()
	content := strings.Join([]string{
		`[INPUT]  [2026-07-20 13:45:00] User 123: hi`,
		`[OUTPUT]  [2026-07-20 13:45:01] User 123: there`,
		`non-matching noise`,
		``,
	}, "\n")
	writeLog(t, ws, "2026-07-20", content)

	date := time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)
	entries, err := ParseDay(ws, date)
	if err != nil {
		t.Fatalf("ParseDay: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if !strings.Contains(entries[1].Raw, "non-matching noise") {
		t.Errorf("continuation line should be in last entry Raw, got %q", entries[1].Raw)
	}
	if entries[0].Level != "INPUT" || entries[0].UserID != 123 {
		t.Errorf("first entry parsed incorrectly: %+v", entries[0])
	}
}
