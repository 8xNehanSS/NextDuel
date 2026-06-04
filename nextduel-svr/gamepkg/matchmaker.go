package gamepkg

import (
	"sync"

	"gorm.io/gorm"
)

type Matchmaking struct {
	mu    sync.Mutex
	Queue []*Player

	db *gorm.DB
}

func NewMatchmaking(db *gorm.DB) *Matchmaking {
	return &Matchmaking{
		Queue: make([]*Player, 0),
		db:    db,
	}
}

func (m *Matchmaking) AddPlayer(player *Player) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Queue = append(m.Queue, player)
	if len(m.Queue) < 2 {
		return nil, false
	}

	p1 := m.Queue[0]
	p2 := m.Queue[1]

	m.Queue = m.Queue[2:]
	room := NewRoom(p1, p2, m.db)

	return room, true
}

func (m *Matchmaking) GetQueue() []*Player {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Queue
}

func (m *Matchmaking) IsPlayerInQueue(player *Player) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Queue {
		if p.ID == player.ID {
			return true
		}
	}
	return false
}

func (m *Matchmaking) RemovePlayer(playerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, p := range m.Queue {
		if p.ID == playerID {
			m.Queue = append(m.Queue[:i], m.Queue[i+1:]...)
			return true
		}
	}
	return false
}

func (m *Matchmaking) GetPlayer(playerID string) (*Player, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Queue {
		if p.ID == playerID {
			return p, true
		}
	}
	return nil, false
}
