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
	Urgency   string    `json:"urgency"`
	CreatedAt time.Time `json:"created_at"`
}

func Init() error {
	homeDir, err := getHomeDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(homeDir, ".opencode-telegram", "data.db")
	dbPath = "file:" + dbPath + "?_busy_timeout=5000&cache=shared"

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
		urgency TEXT DEFAULT 'low',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(notificationsTable); err != nil {
		return err
	}

	if _, err := db.Exec(mailsTable); err != nil {
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

func InsertMail(id string, userID int64, sender, subject, content, urgency string) error {
	_, err := db.Exec(
		"INSERT INTO mails (id, user_id, sender, subject, content, mail_sent, urgency) VALUES (?, ?, ?, ?, ?, FALSE, ?)",
		id, userID, sender, subject, content, urgency,
	)
	return err
}

func GetMail(id string) (*Mail, error) {
	var m Mail
	err := db.QueryRow(
		"SELECT id, user_id, sender, subject, content, mail_sent, urgency, created_at FROM mails WHERE id = ?",
		id,
	).Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.Urgency, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetUnsentMails(userID int64) ([]Mail, error) {
	rows, err := db.Query(
		"SELECT id, user_id, sender, subject, content, mail_sent, urgency, created_at FROM mails WHERE user_id = ? AND mail_sent = FALSE ORDER BY created_at ASC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		var m Mail
		if err := rows.Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.Urgency, &m.CreatedAt); err != nil {
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
		"SELECT id, user_id, sender, subject, content, mail_sent, urgency, created_at FROM mails WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mails []Mail
	for rows.Next() {
		var m Mail
		if err := rows.Scan(&m.ID, &m.UserID, &m.Sender, &m.Subject, &m.Content, &m.MailSent, &m.Urgency, &m.CreatedAt); err != nil {
			return nil, err
		}
		mails = append(mails, m)
	}
	return mails, nil
}
