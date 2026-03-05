package session

import "sync"

type Store struct {
	sessions map[int64]*Session
	mu       sync.RWMutex
}

var store *Store

func NewStore() *Store {
	return &Store{
		sessions: make(map[int64]*Session),
	}
}

func GetStore() *Store {
	if store == nil {
		store = NewStore()
	}
	return store
}

func (s *Store) Get(userID int64) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[userID]
}

func (s *Store) Set(userID int64, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[userID] = session
}

func (s *Store) Delete(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, userID)
}

func (s *Store) GetOrCreate(userID int64, workspace string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[userID]; exists {
		return session
	}

	session := NewSession(userID, workspace)
	s.sessions[userID] = session
	return session
}
