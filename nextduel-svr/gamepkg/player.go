package gamepkg

import (
	"log/slog"
	"nextduel-svr/utils"
	"sync"
	"time"
)

const (
	writeWait      = 5 * time.Second
	sendBufferSize = 64
	maxMessageSize = 4096
)

// Conn is the subset of *websocket.Conn the game needs. It exists so tests
// can substitute an in-memory connection.
type Conn interface {
	WriteJSON(v any) error
	SetWriteDeadline(t time.Time) error
	Close() error
}

// Message is the wire envelope; it matches nex-server's {type, data} shape.
type Message struct {
	Type utils.GameAction `json:"type"`
	Data any              `json:"data"`
}

// Player is one websocket session. gorilla/websocket allows only one
// concurrent writer, so every write goes through a single goroutine fed by a
// buffered channel. Room logic can therefore Send while holding locks without
// ever blocking on the network.
type Player struct {
	ID string

	conn Conn
	send chan Message
	done chan struct{}
	once sync.Once
}

func NewPlayer(id string, conn Conn) *Player {
	// nex-server sets no inbound size limit; cap it on real connections.
	if rl, ok := conn.(interface{ SetReadLimit(int64) }); ok {
		rl.SetReadLimit(maxMessageSize)
	}
	p := &Player{
		ID:   id,
		conn: conn,
		send: make(chan Message, sendBufferSize),
		done: make(chan struct{}),
	}
	go p.writeLoop()
	return p
}

// Send queues a message. If the client is too slow to drain its buffer the
// connection is dropped; the read loop then fails and normal disconnect
// handling runs.
func (p *Player) Send(action utils.GameAction, data any) {
	select {
	case <-p.done:
		return
	default:
	}
	select {
	case p.send <- Message{Type: action, Data: data}:
	default:
		slog.Warn("dropping slow websocket client", "player", p.ID)
		p.Close()
	}
}

// SendError sends a structured error event.
func (p *Player) SendError(err error) {
	p.Send(utils.GameActionError, toErrorPayload(err))
}

// Close stops the writer and closes the underlying connection. Safe to call
// more than once.
func (p *Player) Close() {
	p.once.Do(func() {
		close(p.done)
	})
}

func (p *Player) writeLoop() {
	defer p.conn.Close()
	for {
		select {
		case msg := <-p.send:
			if !p.write(msg) {
				return
			}
		case <-p.done:
			// Flush whatever is already queued (e.g. a final game_over).
			for {
				select {
				case msg := <-p.send:
					if !p.write(msg) {
						return
					}
				default:
					return
				}
			}
		}
	}
}

func (p *Player) write(msg Message) bool {
	_ = p.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := p.conn.WriteJSON(msg); err != nil {
		slog.Debug("websocket write failed", "player", p.ID, "error", err)
		return false
	}
	return true
}
