package session

import "time"

type Session struct {
	UserID      int64
	Agent       string
	Model       string
	Provider    string
	Workspace   string
	OpenCodeID  string
	HistoryPath string
	CreatedAt   time.Time
	LastActive  time.Time
}

func NewSession(userID int64, workspace string) *Session {
	return &Session{
		UserID:     userID,
		Workspace:  workspace,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}
}
