package server

import (
	"modulyn/pkg/models"
	"sync"
)

type store struct {
	mu      sync.RWMutex
	clients map[models.Client]struct{}
}

type Store interface {
	Subscribe(client models.Client)
	Unsubscribe(client models.Client)
}

func NewStore() Store {
	return &store{
		mu:      sync.RWMutex{},
		clients: make(map[models.Client]struct{}),
	}
}

func (s *store) Subscribe(client models.Client) {
	s.mu.Lock()
	s.clients[client] = struct{}{}
	s.mu.Unlock()
}

func (s *store) Unsubscribe(client models.Client) {
	s.mu.Lock()
	delete(s.clients, client)
	s.mu.Unlock()
}
