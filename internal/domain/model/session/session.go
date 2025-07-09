package session

import "time"

type Status string

const (
	Queued Status = "queued"
	Active Status = "active"
	Init   Status = "init"
)

const TTL = time.Hour * 2

type Session struct {
	Token    string `json:"token"`
	Status   Status `json:"status"`
	Entrance int64  `json:"entrance"`
}

func NewSession(token string, status Status, entrance int64) Session {
	return Session{
		Token:    token,
		Status:   status,
		Entrance: entrance,
	}
}

func (s *Session) SetStatus(status Status) {
	s.Status = status
}

func (s *Session) IsQueued() bool {
	return s.Status == Queued
}

func (s *Session) IsActive() bool {
	return s.Status == Active
}
