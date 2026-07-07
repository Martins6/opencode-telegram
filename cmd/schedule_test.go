package cmd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/spf13/cobra"
)

func TestParseScheduleAtOnceWithTimezone(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	tests := []struct {
		expr       string
		wantHour   int
		wantMinute int
	}{
		{"at 09:00", 9, 0},
		{"once 14:30", 14, 30},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			next, err := parseSchedule(tt.expr, loc)
			if err != nil {
				t.Fatalf("parseSchedule(%q) error: %v", tt.expr, err)
			}
			if next.Hour() != tt.wantHour {
				t.Errorf("hour = %d, want %d", next.Hour(), tt.wantHour)
			}
			if next.Minute() != tt.wantMinute {
				t.Errorf("minute = %d, want %d", next.Minute(), tt.wantMinute)
			}
			if got := next.Location().String(); got != "America/Sao_Paulo" {
				t.Errorf("location = %q, want America/Sao_Paulo", got)
			}
		})
	}
}

func TestParseScheduleCronWithTimezone(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	next, err := parseSchedule("0 9 * * *", loc)
	if err != nil {
		t.Fatalf("parseSchedule error: %v", err)
	}
	if next.Hour() != 9 {
		t.Errorf("hour = %d, want 9 (in BRT)", next.Hour())
	}
	if next.Location().String() != "America/Sao_Paulo" {
		t.Errorf("location = %q, want America/Sao_Paulo", next.Location().String())
	}
	utcHour := next.UTC().Hour()
	if utcHour != 12 {
		t.Errorf("UTC hour = %d, want 12 (09:00 BRT = 12:00 UTC)", utcHour)
	}
}

func TestParseScheduleNilLocation(t *testing.T) {
	_, err := parseSchedule("at 09:00", nil)
	if !errors.Is(err, config.ErrTimezoneNotConfigured) {
		t.Errorf("got error %v, want ErrTimezoneNotConfigured", err)
	}
}

func TestParseScheduleAtRollsForwardIfPast(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	expr := "at 00:01"
	next, err := parseSchedule(expr, loc)
	if err != nil {
		t.Fatalf("parseSchedule error: %v", err)
	}
	if !next.After(time.Now()) {
		t.Errorf("expected next run to be in the future, got %v", next)
	}
	if next.Location().String() != "America/Sao_Paulo" {
		t.Errorf("location = %q, want America/Sao_Paulo", next.Location().String())
	}
}

func TestParseScheduleInAndNow(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("failed to load location: %v", err)
	}

	before := time.Now()
	next, err := parseSchedule("in 30m", loc)
	if err != nil {
		t.Fatalf("parseSchedule(in 30m) error: %v", err)
	}
	expectedMin := before.Add(30 * time.Minute)
	if next.Before(expectedMin.Add(-time.Second)) || next.After(expectedMin.Add(time.Second)) {
		t.Errorf("got %v, want ~%v", next, expectedMin)
	}

	next, err = parseSchedule("now + 1h", loc)
	if err != nil {
		t.Fatalf("parseSchedule(now + 1h) error: %v", err)
	}
	expectedMin = before.Add(time.Hour)
	if next.Before(expectedMin.Add(-time.Second)) || next.After(expectedMin.Add(time.Second)) {
		t.Errorf("got %v, want ~%v", next, expectedMin)
	}
}

func TestRequireTimezoneBlocksWithoutConfig(t *testing.T) {
	orig := config.Get()
	defer config.SetForTest(orig)

	config.SetForTest(&config.Config{Bot: config.BotConfig{Timezone: ""}})

	fakeCmd := &cobra.Command{Use: "add"}
	err := requireTimezone(fakeCmd)
	if err == nil {
		t.Fatal("expected gate error, got nil")
	}
	if !strings.Contains(err.Error(), "timezone not configured") {
		t.Errorf("error = %q, want it to contain 'timezone not configured'", err.Error())
	}
}

func TestRequireTimezoneAllowsSetAndParent(t *testing.T) {
	orig := config.Get()
	defer config.SetForTest(orig)

	config.SetForTest(&config.Config{Bot: config.BotConfig{Timezone: ""}})

	if err := requireTimezone(&cobra.Command{Use: "set"}); err != nil {
		t.Errorf("set subcommand should be exempt, got error: %v", err)
	}
	if err := requireTimezone(&cobra.Command{Use: "schedule"}); err != nil {
		t.Errorf("parent command should be exempt, got error: %v", err)
	}
}

func TestRequireTimezoneAllowsWhenConfigured(t *testing.T) {
	orig := config.Get()
	defer config.SetForTest(orig)

	config.SetForTest(&config.Config{Bot: config.BotConfig{Timezone: "America/Sao_Paulo"}})

	if err := requireTimezone(&cobra.Command{Use: "add"}); err != nil {
		t.Errorf("add subcommand should pass when timezone is configured, got error: %v", err)
	}
}
