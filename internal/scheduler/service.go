package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/martins6/opencode-telegram/internal/logger"
	"github.com/robfig/cron/v3"
)

type SchedulerService struct {
	bot           *bot.Bot
	ctx           context.Context
	cancel        context.CancelFunc
	workspacePath string
	userID        int64
	cron          *cron.Cron
}

func StartScheduler(ctx context.Context, b *bot.Bot, workspacePath string) error {
	if err := database.Init(workspacePath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	userID := config.GetAllowedUserChatID()

	s := &SchedulerService{
		bot:           b,
		ctx:           ctx,
		workspacePath: workspacePath,
		userID:        userID,
		cron:          cron.New(),
	}
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.cron.Start()
	go s.run()

	logger.LogDebug("Scheduler service started")
	return nil
}

func (s *SchedulerService) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer s.cron.Stop()

	for {
		select {
		case <-s.ctx.Done():
			logger.LogDebug("Scheduler service stopped")
			return
		case <-ticker.C:
			s.processDueTasks()
		}
	}
}

func (s *SchedulerService) processDueTasks() {
	userID := config.GetAllowedUserChatID()
	if userID == 0 {
		logger.LogDebug("Scheduler: cannot process tasks - user chat ID not resolved. Please message the bot first.")
		return
	}

	tasks, err := database.GetDueScheduledTasks(userID)
	if err != nil {
		logger.LogDebug("Scheduler: failed to get due tasks for user %d: %v", userID, err)
		return
	}

	for _, task := range tasks {
		s.executeTask(task, userID)
	}
}

func (s *SchedulerService) executeTask(task database.ScheduledTask, userID int64) {
	logger.LogDebug("Scheduler: executing task %d: %s", task.ID, task.Command)

	now := time.Now()
	var nextRun time.Time

	isOneTime := strings.HasPrefix(task.ScheduleExpr, "once ") ||
		strings.HasPrefix(task.ScheduleExpr, "at ") ||
		strings.HasPrefix(task.ScheduleExpr, "now ") ||
		strings.HasPrefix(task.ScheduleExpr, "in ")

	if isOneTime {
		database.UpdateScheduledTaskStatus(task.ID, "completed")
	} else {
		specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		expr, err := specParser.Parse(task.ScheduleExpr)
		if err != nil {
			logger.LogDebug("Scheduler: failed to parse cron expression: %v", err)
			database.UpdateScheduledTaskStatus(task.ID, "failed")
		} else {
			nextRun = expr.Next(now)
			database.UpdateScheduledTaskRun(task.ID, &now, &nextRun)
		}
	}

	cmdStr := strings.TrimSpace(task.Command)

	isNotifyCommand := strings.HasPrefix(cmdStr, "opencode-telegram notify") ||
		strings.HasPrefix(cmdStr, "notify")

	if isNotifyCommand {
		var msg string
		if strings.HasPrefix(cmdStr, "opencode-telegram notify -m ") {
			msg = strings.TrimPrefix(cmdStr, "opencode-telegram notify -m ")
		} else if strings.HasPrefix(cmdStr, "opencode-telegram notify ") {
			msg = strings.TrimPrefix(cmdStr, "opencode-telegram notify ")
		} else if strings.HasPrefix(cmdStr, "notify -m ") {
			msg = strings.TrimPrefix(cmdStr, "notify -m ")
		} else if strings.HasPrefix(cmdStr, "notify ") {
			msg = strings.TrimPrefix(cmdStr, "notify ")
		} else if strings.Contains(cmdStr, " -m ") {
			parts := strings.Split(cmdStr, " -m ")
			if len(parts) > 1 {
				msg = parts[1]
			}
		}
		msg = strings.Trim(msg, "\"")
		database.InsertNotification(userID, msg)
		logger.LogDebug("Scheduler: notification created for task %d", task.ID)
		return
	}

	if strings.HasPrefix(cmdStr, "opencode-telegram mail send") {
		args := parseMailCommandArgs(cmdStr)
		sender := args["--sender"]
		subject := args["--subject"]
		content := args["--content"]

		if sender == "" {
			sender = "scheduler@system"
		}
		if subject == "" {
			subject = "Scheduled Task"
		}
		mailID := generateUUID()
		database.InsertMail(mailID, userID, sender, subject, content)
		logger.LogDebug("Scheduler: mail created for task %d (mailID: %s, sender: %s, subject: %s)", task.ID, mailID, sender, subject)
		return
	}

	workingDir := task.WorkingDir
	if workingDir == "" {
		workingDir = s.workspacePath
	}

	if workingDir == "" {
		homeDir, _ := os.UserHomeDir()
		workingDir = filepath.Join(homeDir, ".opencode-telegram")
	}

	cmd := getCommandForTask(cmdStr)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()

	if err != nil {
		if task.OnFailure == "notify" {
			msg := fmt.Sprintf("Scheduled task failed: %s\nError: %s", task.Command, err.Error())
			database.InsertNotification(userID, msg)
		}
		logger.LogDebug("Scheduler: task %d failed: %v", task.ID, err)
	} else {
		if task.OnSuccess == "mail" {
			subject := fmt.Sprintf("Scheduled Task: %s", task.Command)
			if len(subject) > 50 {
				subject = subject[:50] + "..."
			}
			mailID := generateUUID()
			database.InsertMail(
				mailID,
				userID,
				"scheduler@system",
				subject,
				string(output),
			)
			logger.LogDebug("Scheduler: task %d completed, mail created (mailID: %s, subject: %s)", task.ID, mailID, subject)
		} else if task.OnSuccess == "notify" {
			msg := fmt.Sprintf("Scheduled task completed: %s\nOutput: %s", task.Command, string(output))
			database.InsertNotification(userID, msg)
			logger.LogDebug("Scheduler: task %d completed, notification created", task.ID)
		} else {
			logger.LogDebug("Scheduler: task %d completed successfully", task.ID)
		}
	}
}

func (s *SchedulerService) Stop() {
	if s != nil && s.cancel != nil {
		s.cancel()
	}
	logger.LogDebug("Stopping scheduler service...")
}

func getCommandForTask(cmdStr string) *exec.Cmd {
	if strings.HasPrefix(cmdStr, "opencode-telegram ") {
		execPath, err := os.Executable()
		if err == nil {
			args := strings.Fields(cmdStr)[1:]
			return exec.Command(execPath, args...)
		}
	}
	return exec.Command("sh", "-c", cmdStr)
}

func generateUUID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(i * 17 % 256)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func parseMailCommandArgs(cmdStr string) map[string]string {
	args := make(map[string]string)
	remaining := strings.TrimPrefix(cmdStr, "opencode-telegram mail send")
	remaining = strings.TrimSpace(remaining)

	flagNames := []string{"--sender", "--subject", "--content"}

	for _, flag := range flagNames {
		idx := strings.Index(remaining, flag)
		if idx == -1 {
			continue
		}

		valueStart := idx + len(flag)
		valueEnd := len(remaining)

		nextFlagIdx := len(remaining)
		for _, nextFlag := range flagNames {
			if nextFlag == flag {
				continue
			}
			ni := strings.Index(remaining[valueStart:], nextFlag)
			if ni != -1 && valueStart+ni < nextFlagIdx {
				nextFlagIdx = valueStart + ni
			}
		}

		valueEnd = nextFlagIdx

		value := strings.TrimSpace(remaining[valueStart:valueEnd])
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		} else if strings.HasPrefix(value, "\"") {
			value = value[1:]
		}

		args[flag] = value
	}

	return args
}
