package gamepkg

// Matchmaking is a FIFO queue that pairs players two at a time.
// It is not safe for concurrent use; GameManager guards it with its mutex.
type Matchmaking struct {
	queue []*Player
}

func NewMatchmaking() *Matchmaking {
	return &Matchmaking{}
}

// Enqueue adds p. If that completes a pair, the pair is removed from the
// queue and returned; otherwise p's 1-based queue position is returned.
func (m *Matchmaking) Enqueue(p *Player) (pair []*Player, position int) {
	m.queue = append(m.queue, p)
	if len(m.queue) < 2 {
		return nil, len(m.queue)
	}
	pair = []*Player{m.queue[0], m.queue[1]}
	m.queue = m.queue[2:]
	return pair, 0
}

// Remove drops this exact session from the queue.
func (m *Matchmaking) Remove(p *Player) bool {
	for i, q := range m.queue {
		if q == p {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			return true
		}
	}
	return false
}

func (m *Matchmaking) Contains(playerID string) bool {
	for _, q := range m.queue {
		if q.ID == playerID {
			return true
		}
	}
	return false
}

func (m *Matchmaking) Len() int {
	return len(m.queue)
}

// Drain empties the queue and returns who was in it.
func (m *Matchmaking) Drain() []*Player {
	out := m.queue
	m.queue = nil
	return out
}
