package broadcaster

import (
	"sync"
	"time"
)

type LogEvent struct {
	Time          time.Time `json:"time"`
	SourceIP      string    `json:"source_ip"`
	StatusCode    int       `json:"status_code"`
	RequestedFile string    `json:"requested_file"`
	Method        string    `json:"method"`
	UserAgent     string    `json:"user_agent"`
	Message       string    `json:"message,omitempty"`
}

type Broadcaster struct {
	subscribers []chan LogEvent
	mu          sync.Mutex
}

func New() *Broadcaster {
	return &Broadcaster{
		subscribers: make([]chan LogEvent, 0),
	}
}

func (b *Broadcaster) Subscribe() chan LogEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan LogEvent, 100) // Buffer to prevent blocking
	b.subscribers = append(b.subscribers, ch)
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan LogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, sub := range b.subscribers {
		if sub == ch {
			close(sub)
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			return
		}
	}
}

func (b *Broadcaster) Broadcast(event LogEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Drop event if subscriber is too slow
		}
	}
}
