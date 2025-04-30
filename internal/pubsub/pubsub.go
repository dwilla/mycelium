package pubsub

import (
	"sync"
)

type TypingEvent struct {
	UserID   string
	Username string
	Channel  string `json:"channel"`
	Message  string `json:"message"`
}

type PubSub struct {
	subscribers map[string]map[chan TypingEvent]struct{}
	mu          sync.RWMutex
}

func New() *PubSub {
	return &PubSub{
		subscribers: make(map[string]map[chan TypingEvent]struct{}),
	}
}

func (ps *PubSub) Subscribe(channelID string) chan TypingEvent {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan TypingEvent)
	if ps.subscribers[channelID] == nil {
		ps.subscribers[channelID] = make(map[chan TypingEvent]struct{})
	}
	ps.subscribers[channelID][ch] = struct{}{}
	return ch
}

func (ps *PubSub) Unsubscribe(channelID string, ch chan TypingEvent) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, ok := ps.subscribers[channelID]; ok {
		close(ch)
		delete(subs, ch)
		if len(subs) == 0 {
			delete(ps.subscribers, channelID)
		}
	}
}

func (ps *PubSub) Publish(channelID string, event TypingEvent) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subs, exists := ps.subscribers[channelID]; exists {
		for ch := range subs {
			select {
			case ch <- event:
			default:
				// Skip
			}
		}
	}
}
