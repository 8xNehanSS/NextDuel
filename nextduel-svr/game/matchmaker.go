package game

import (
	"nextduel-svr/models"
	"sync"
)

type Matchmaking struct {
	mu    sync.Mutex
	Queue []*models.Player
}

func NewMatchmaking() *Matchmaking {
	return &Matchmaking{
		Queue: make([]*models.Player, 0),
	}
}

func (m *Matchmaking) AddPlayer(player *models.Player) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Queue = append(m.Queue, player)
	if len(m.Queue) < 2 {
		return nil, false
	}

	p1 := m.Queue[0]
	p2 := m.Queue[1]

	m.Queue = m.Queue[2:]
	room := NewRoom(p1, p2)

	return room, true
}

func (m *Matchmaking) GetQueue() []*models.Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Queue
}

func (m *Matchmaking) IsPlayerInQueue(player *models.Player) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Queue {
		if p.ID == player.ID {
			return true
		}
	}
	return false
}
