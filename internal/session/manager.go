package session

import (
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
	store *Store
}

var manager *Manager

func NewManager() *Manager {
	return &Manager{
		store: NewStore(),
	}
}

func GetManager() *Manager {
	if manager == nil {
		manager = NewManager()
	}
	return manager
}

func (m *Manager) GetSession(userID int64, workspace string) (*Session, error) {
	session := m.store.GetOrCreate(userID, workspace)

	if session.HistoryPath == "" {
		historyPath := filepath.Join(workspace, "conversations", fmt.Sprintf("%d.json", userID))
		if err := os.MkdirAll(filepath.Dir(historyPath), 0755); err != nil {
			return nil, err
		}
		session.HistoryPath = historyPath
	}

	return session, nil
}

func (m *Manager) UpdateSession(session *Session) {
	m.store.Set(session.UserID, session)
}

func (m *Manager) ResetSession(userID int64) {
	m.store.Delete(userID)
}

func (m *Manager) GetAgent(session *Session) string {
	if session == nil {
		return ""
	}
	return session.Agent
}

func (m *Manager) SetAgent(session *Session, agent string) {
	if session != nil {
		session.Agent = agent
	}
}

func (m *Manager) GetModel(session *Session) string {
	if session == nil {
		return ""
	}
	return session.Model
}

func (m *Manager) SetModel(session *Session, model string) {
	if session != nil {
		session.Model = model
	}
}

func (m *Manager) GetProvider(session *Session) string {
	if session == nil {
		return ""
	}
	return session.Provider
}

func (m *Manager) SetProvider(session *Session, provider string) {
	if session != nil {
		session.Provider = provider
	}
}

func (m *Manager) GetWorkspace(session *Session) string {
	if session == nil {
		return ""
	}
	return session.Workspace
}

func (m *Manager) SetWorkspace(session *Session, workspace string) {
	if session != nil {
		session.Workspace = workspace
	}
}
