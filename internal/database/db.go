package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type Notification struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Message     string    `json:"message"`
	MessageSent bool      `json:"message_sent"`
	CreatedAt   time.Time `json:"created_at"`
}

type Mail struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Sender    string    `json:"sender"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	MailSent  bool      `json:"mail_sent"`
	CreatedAt time.Time `json:"created_at"`
}

type ScheduledTask struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	ScheduleExpr string     `json:"schedule_expr"`
	Command      string     `json:"command"`
	WorkingDir   string     `json:"working_dir"`
	OnSuccess    string     `json:"on_success"`
	OnFailure    string     `json:"on_failure"`
	Status       string     `json:"status"`
	LastRun      *time.Time `json:"last_run"`
	NextRun      *time.Time `json:"next_run"`
	CreatedAt    time.Time  `json:"created_at"`
}

func Init(workspacePath string) error {
	homeDir := workspacePath
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
		homeDir = filepath.Join(homeDir, ".opencode-telegram")
	}

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, "data.db")

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

func createTables() error {
	notificationsTable := `
	CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		message TEXT NOT NULL,
		message_sent BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	mailsTable := `
	CREATE TABLE IF NOT EXISTS mails (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		sender TEXT NOT NULL,
		subject TEXT NOT NULL,
		content TEXT NOT NULL,
		mail_sent BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	scheduledTasksTable := `
	CREATE TABLE IF NOT EXISTS scheduled_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		schedule_expr TEXT NOT NULL,
		command TEXT NOT NULL,
		working_dir TEXT,
		on_success TEXT DEFAULT 'mail',
		on_failure TEXT DEFAULT 'notify',
		status TEXT DEFAULT 'active',
		last_run DATETIME,
		next_run DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	resolvedChatIDsTable := `
	CREATE TABLE IF NOT EXISTS resolved_chat_ids (
		username TEXT PRIMARY KEY,
		chat_id INTEGER NOT NULL
	);`

	if _, err := db.Exec(notificationsTable); err != nil {
		return err
	}

	if _, err := db.Exec(mailsTable); err != nil {
		return err
	}

	if _, err := db.Exec(scheduledTasksTable); err != nil {
		return err
	}

	if _, err := db.Exec(resolvedChatIDsTable); err != nil {
		return err
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func getHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".opencode-telegram"), nil
}

func InsertNotification(userID int64, message string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO notifications (user_id, message, message_sent) VALUES (?, ?, FALSE)",
		userID, message,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUnsentNotifications(userID int64) ([]Notification, error) {
	rows, err := db.Query(
		"SELECT id, user_id, message, message_sent, created_at FROM notifications WHERE user_id = ? AND message_sent = FALSE ORDER BY created_at ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Message, &n.MessageSent, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

func MarkNotificationSent(id int64) error {
	_, err := db.Exec("UPDATE notifications SET message_sent = TRUE WHERE id = ?", id)
	return err
}

func InsertMail(id string, userID int64, sender, subject, content string) error {
	_, err := db.Exec(
		"INSERT INTO mails (id, user_id, sender, subject, content, mail_sent) VALUES (?, ?, ?, ?, ?, FALSE)",
		id, userID, sender, subject, content,
	)
	return err
}

func GetMail(id string) (*Mail, error) {
	var m Mail
	err := db.QueryRow(
		"SELECT id, user_id, sender, subject, content, mail_sent, created_at FROM mails WHERE id = ?",
		id,
	).Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetUnsentMails(userID int64) ([]Mail, error) {
	rows, err := db.Query(
		"SELECT id, user_id, sender, subject, content, mail_sent, created_at FROM mails WHERE user_id = ? AND mail_sent = FALSE ORDER BY created_at ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		var m Mail
		if err := rows.Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.CreatedAt); err != nil {
			return nil, err
		}
		mails = append(mails, m)
	}
	return mails, nil
}

func MarkMailSent(id string) error {
	_, err := db.Exec("UPDATE mails SET mail_sent = TRUE WHERE id = ?", id)
	return err
}

func ListMails(userID int64) ([]Mail, error) {
	rows, err := db.Query(
		"SELECT id, user_id, sender, subject, content, mail_sent, created_at FROM mails WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		var m Mail
		if err := rows.Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.CreatedAt); err != nil {
			return nil, err
		}
		mails = append(mails, m)
	}
	return mails, nil
}

func InsertScheduledTask(userID int64, scheduleExpr, command, workingDir, onSuccess, onFailure string, nextRun *time.Time) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO scheduled_tasks (user_id, schedule_expr, command, working_dir, on_success, on_failure, status, next_run) VALUES (?, ?, ?, ?, ?, ?, 'active', ?)",
		userID, scheduleExpr, command, workingDir, onSuccess, onFailure, nextRun,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetScheduledTask(id int64) (*ScheduledTask, error) {
	var t ScheduledTask
	err := db.QueryRow(
		"SELECT id, user_id, schedule_expr, command, working_dir, on_success, on_failure, status, last_run, next_run, created_at FROM scheduled_tasks WHERE id = ?",
		id,
	).Scan(&t.ID, &t.UserID, &t.ScheduleExpr, &t.Command, &t.WorkingDir, &t.OnSuccess, &t.OnFailure, &t.Status, &t.LastRun, &t.NextRun, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetDueScheduledTasks(userID int64) ([]ScheduledTask, error) {
	rows, err := db.Query(
		"SELECT id, user_id, schedule_expr, command, working_dir, on_success, on_failure, status, last_run, next_run, created_at FROM scheduled_tasks WHERE user_id = ? AND status = 'active' AND next_run <= datetime('now', 'localtime') ORDER BY next_run ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ScheduledTask
	for rows.Next() {
		var t ScheduledTask
		if err := rows.Scan(&t.ID, &t.UserID, &t.ScheduleExpr, &t.Command, &t.WorkingDir, &t.OnSuccess, &t.OnFailure, &t.Status, &t.LastRun, &t.NextRun, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func ListScheduledTasks(userID int64) ([]ScheduledTask, error) {
	rows, err := db.Query(
		"SELECT id, user_id, schedule_expr, command, working_dir, on_success, on_failure, status, last_run, next_run, created_at FROM scheduled_tasks WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ScheduledTask
	for rows.Next() {
		var t ScheduledTask
		if err := rows.Scan(&t.ID, &t.UserID, &t.ScheduleExpr, &t.Command, &t.WorkingDir, &t.OnSuccess, &t.OnFailure, &t.Status, &t.LastRun, &t.NextRun, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func UpdateScheduledTaskStatus(id int64, status string) error {
	_, err := db.Exec("UPDATE scheduled_tasks SET status = ? WHERE id = ?", status, id)
	return err
}

func UpdateScheduledTaskRun(id int64, lastRun, nextRun *time.Time) error {
	_, err := db.Exec("UPDATE scheduled_tasks SET last_run = ?, next_run = ? WHERE id = ?", lastRun, nextRun, id)
	return err
}

func DeleteScheduledTask(id int64) error {
	_, err := db.Exec("DELETE FROM scheduled_tasks WHERE id = ?", id)
	return err
}

func SetResolvedChatID(username string, chatID int64) error {
	_, err := db.Exec(
		"INSERT OR REPLACE INTO resolved_chat_ids (username, chat_id) VALUES (?, ?)",
		username, chatID,
	)
	return err
}

func GetResolvedChatID(username string) (int64, error) {
	var chatID int64
	err := db.QueryRow(
		"SELECT chat_id FROM resolved_chat_ids WHERE username = ?",
		username,
	).Scan(&chatID)
	if err != nil {
		return 0, err
	}
	return chatID, nil
}
