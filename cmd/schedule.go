package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/database"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scheduleCmd         string
	scheduleCommand     string
	scheduleDir         string
	scheduleOnSucc      string
	scheduleOnFail      string
	scheduleSetTimezone string
)

func requireTimezone(cmd *cobra.Command) error {
	if cmd.Name() == "schedule" || cmd.Name() == "set" {
		return nil
	}
	if _, err := config.GetLocation(); err != nil {
		return fmt.Errorf("timezone not configured. Run 'acolyte schedule set --timezone <IANA>' first")
	}
	return nil
}

var scheduleCmdMain = &cobra.Command{
	Use:   "schedule",
	Short: "Manage scheduled tasks",
	Long:  "Schedule shell commands to run automatically at specified times.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if _, err := config.Load(""); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		}

		workspacePath := viper.GetString("workspace.path")
		if workspacePath == "" {
			homeDir, _ := os.UserHomeDir()
			workspacePath = filepath.Join(homeDir, ".acolyte")
		}
		if err := database.Init(workspacePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize database: %v\n", err)
		}

		return requireTimezone(cmd)
	},
}

var scheduleSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set the timezone used by the scheduler (one-time setup)",
	Long: `Set the timezone used by the scheduler. This is a one-time setup required
before any other 'schedule' subcommand will work.

Example:
  acolyte schedule set --timezone America/Sao_Paulo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if scheduleSetTimezone == "" {
			return fmt.Errorf("timezone is required (use --timezone, e.g. America/Sao_Paulo)")
		}
		loc, err := time.LoadLocation(scheduleSetTimezone)
		if err != nil {
			return fmt.Errorf("invalid timezone: %w", err)
		}

		previousZone := ""
		if cfg := config.Get(); cfg != nil {
			previousZone = cfg.Bot.Timezone
		}

		viper.Set("bot.timezone", scheduleSetTimezone)

		configPath := viper.ConfigFileUsed()
		if configPath == "" {
			homeDir, _ := os.UserHomeDir()
			configPath = filepath.Join(homeDir, ".acolyte", "config.toml")
		}
		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		if _, err := config.Load(configPath); err != nil {
			return fmt.Errorf("failed to reload config: %w", err)
		}

		now := time.Now().In(loc)
		fmt.Printf("Timezone set to %s. Current time in that zone: %s\n", scheduleSetTimezone, now.Format(time.RFC3339))

		if previousZone != "" && previousZone != scheduleSetTimezone {
			userID := config.GetAllowedUserChatID()
			if userID != 0 {
				tasks, err := database.ListScheduledTasks(userID)
				if err == nil && len(tasks) > 0 {
					fmt.Printf("Note: %d existing scheduled tasks will be re-interpreted under the new zone.\n", len(tasks))
				}
			}
		}
		return nil
	},
}

var scheduleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new scheduled task",
	Long: `Add a new scheduled task that executes a shell command.
		
Examples:
  # Run once in 30 minutes
  acolyte schedule add --schedule "in 30m" --command "echo hello"
  
  # Run at specific time
  acolyte schedule add --schedule "at 09:00" --command "backup.sh"
  
  # Run daily at 9am (cron)
  acolyte schedule add --schedule "0 9 * * *" --command "daily-task.sh"
  
  # Run every 15 minutes (cron)
  acolyte schedule add --schedule "*/15 * * * *" --command "check-status.sh"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if scheduleCmd == "" {
			return fmt.Errorf("schedule expression is required (use --schedule)")
		}
		if scheduleCommand == "" {
			return fmt.Errorf("command is required (use --command)")
		}

		userID := config.GetAllowedUserChatID()
		if userID == 0 {
			return fmt.Errorf("please send a message to the bot first to register your chat ID")
		}

		loc, err := config.GetLocation()
		if err != nil {
			return fmt.Errorf("failed to resolve timezone: %w", err)
		}
		nextRun, err := parseSchedule(scheduleCmd, loc)
		if err != nil {
			return fmt.Errorf("failed to parse schedule: %w", err)
		}

		onSuccess := scheduleOnSucc
		if onSuccess == "" {
			onSuccess = "mail"
		}
		if onSuccess != "mail" && onSuccess != "notify" {
			return fmt.Errorf("on-success must be 'mail' or 'notify'")
		}

		onFailure := scheduleOnFail
		if onFailure == "" {
			onFailure = "notify"
		}
		if onFailure != "notify" && onFailure != "ignore" {
			return fmt.Errorf("on-failure must be 'notify' or 'ignore'")
		}

		dir := scheduleDir
		if dir == "" {
			homeDir, _ := os.UserHomeDir()
			dir = filepath.Join(homeDir, ".acolyte")
		}

		id, err := database.InsertScheduledTask(userID, scheduleCmd, scheduleCommand, dir, onSuccess, onFailure, nextRun)
		if err != nil {
			return fmt.Errorf("failed to insert scheduled task: %w", err)
		}

		fmt.Printf("Scheduled task created with ID: %d\n", id)
		fmt.Printf("Command: %s\n", scheduleCommand)
		fmt.Printf("Schedule: %s\n", scheduleCmd)
		fmt.Printf("Next run: %s\n", nextRun.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var scheduleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scheduled tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := config.GetAllowedUserChatID()
		if userID == 0 {
			return fmt.Errorf("please send a message to the bot first to register your chat ID")
		}

		tasks, err := database.ListScheduledTasks(userID)
		if err != nil {
			return fmt.Errorf("failed to list scheduled tasks: %w", err)
		}

		if len(tasks) == 0 {
			fmt.Println("No scheduled tasks found.")
			return nil
		}

		fmt.Println("ID\t Schedule\t\t\t Command\t\t\t\t Status\t Next Run")
		fmt.Println("---------------------------------------------------------------------------------------------------")
		for _, t := range tasks {
			nextRun := "N/A"
			if t.NextRun != nil {
				nextRun = t.NextRun.Format("2006-01-02 15:04")
			}
			cmdShort := t.Command
			if len(cmdShort) > 30 {
				cmdShort = cmdShort[:30] + "..."
			}
			fmt.Printf("%d\t %-20s\t %-30s\t %s\t %s\n", t.ID, t.ScheduleExpr, cmdShort, t.Status, nextRun)
		}
		return nil
	},
}

var scheduleDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a scheduled task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		_, err := fmt.Sscanf(args[0], "%d", &id)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		err = database.DeleteScheduledTask(id)
		if err != nil {
			return fmt.Errorf("failed to delete scheduled task: %w", err)
		}

		fmt.Printf("Scheduled task %d deleted.\n", id)
		return nil
	},
}

var scheduleRunCmd = &cobra.Command{
	Use:   "run [id]",
	Short: "Run a scheduled task immediately",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		_, err := fmt.Sscanf(args[0], "%d", &id)
		if err != nil {
			return fmt.Errorf("invalid task ID: %w", err)
		}

		task, err := database.GetScheduledTask(id)
		if err != nil {
			return fmt.Errorf("failed to get scheduled task: %w", err)
		}

		fmt.Printf("Running task %d: %s\n", id, task.Command)

		workingDir := task.WorkingDir
		if workingDir == "" {
			homeDir, _ := os.UserHomeDir()
			workingDir = filepath.Join(homeDir, ".acolyte")
		}

		execCmd := exec.Command("sh", "-c", task.Command)
		execCmd.Dir = workingDir
		execCmd.Env = os.Environ()

		output, err := execCmd.CombinedOutput()

		userID := config.GetAllowedUserChatID()
		if userID == 0 {
			return fmt.Errorf("please send a message to the bot first to register your chat ID")
		}

		if err != nil {
			if task.OnFailure == "notify" {
				msg := fmt.Sprintf("Task failed: %s\nError: %s", task.Command, err.Error())
				database.InsertNotification(userID, msg)
			}
			fmt.Printf("Task failed: %v\n%s\n", err, output)
		} else {
			if task.OnSuccess == "mail" {
				subject := fmt.Sprintf("Scheduled Task: %s", task.Command)
				database.InsertMail(
					generateUUID(),
					userID,
					"scheduler@system",
					subject,
					string(output),
				)
				fmt.Printf("Output sent to mail.\n")
			} else if task.OnSuccess == "notify" {
				msg := fmt.Sprintf("Task completed: %s\nOutput: %s", task.Command, string(output))
				database.InsertNotification(userID, msg)
			}
			fmt.Printf("Task completed successfully.\n%s\n", output)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scheduleCmdMain)
	scheduleCmdMain.AddCommand(scheduleSetCmd)
	scheduleCmdMain.AddCommand(scheduleAddCmd)
	scheduleCmdMain.AddCommand(scheduleListCmd)
	scheduleCmdMain.AddCommand(scheduleDeleteCmd)
	scheduleCmdMain.AddCommand(scheduleRunCmd)

	scheduleSetCmd.Flags().StringVar(&scheduleSetTimezone, "timezone", "", "IANA timezone, e.g. America/Sao_Paulo")

	scheduleAddCmd.Flags().StringVarP(&scheduleCmd, "schedule", "s", "", "Schedule expression (cron, 'at HH:MM', 'in 30m', 'once HH:MM')")
	scheduleAddCmd.Flags().StringVarP(&scheduleCommand, "command", "c", "", "Shell command to execute")
	scheduleAddCmd.Flags().StringVarP(&scheduleDir, "dir", "d", "", "Working directory (default: workspace)")
	scheduleAddCmd.Flags().StringVar(&scheduleOnSucc, "on-success", "mail", "Action on success: mail or notify")
	scheduleAddCmd.Flags().StringVar(&scheduleOnFail, "on-failure", "notify", "Action on failure: notify or ignore")
}

func parseSchedule(expr string, loc *time.Location) (*time.Time, error) {
	if loc == nil {
		return nil, config.ErrTimezoneNotConfigured
	}
	expr = strings.TrimSpace(expr)

	inPattern := regexp.MustCompile(`(?i)^in\s+(\d+)([smh])$`)
	atPattern := regexp.MustCompile(`(?i)^at\s+(\d{1,2}:\d{2})`)
	oncePattern := regexp.MustCompile(`(?i)^once\s+(\d{1,2}:\d{2})`)
	nowPattern := regexp.MustCompile(`(?i)^now\s*\+?\s*(\d+)([smh])$`)

	if matches := inPattern.FindStringSubmatch(expr); len(matches) == 3 {
		duration, err := time.ParseDuration(matches[1] + matches[2])
		if err != nil {
			return nil, err
		}
		next := time.Now().Add(duration)
		return &next, nil
	}

	if matches := nowPattern.FindStringSubmatch(expr); len(matches) == 3 {
		duration, err := time.ParseDuration(matches[1] + matches[2])
		if err != nil {
			return nil, err
		}
		next := time.Now().Add(duration)
		return &next, nil
	}

	if matches := atPattern.FindStringSubmatch(expr); len(matches) == 2 {
		now := time.Now().In(loc)
		t, err := time.Parse("15:04", matches[1])
		if err != nil {
			return nil, err
		}
		next := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return &next, nil
	}

	if matches := oncePattern.FindStringSubmatch(expr); len(matches) == 2 {
		now := time.Now().In(loc)
		t, err := time.Parse("15:04", matches[1])
		if err != nil {
			return nil, err
		}
		next := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return &next, nil
	}

	cronPattern := regexp.MustCompile(`^[\d\*\/\-\,]+ [\d\*\/\-\,]+ [\d\*\/\-\,]+ [\d\*\/\-\,]+ [\d\*\/\-\,]+$`)
	if cronPattern.MatchString(expr) {
		specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		parsed, err := specParser.Parse(expr)
		if err != nil {
			return nil, err
		}
		next := parsed.Next(time.Now().In(loc))
		return &next, nil
	}

	return nil, fmt.Errorf("invalid schedule expression: %s", expr)
}
