package pubsub

import (
	"sync"
)

type TypingEvent struct {
	UserID    string
	ChannelID string
	Message   string
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

	ch := make(chan TypingEvent, 10)
	if _, exists := ps.subscribers[channelID]; !exists {
		ps.subscribers[channelID] = make(map[chan TypingEvent]struct{})
	}
	ps.subscribers[channelID][ch] = struct{}{}
	return ch
}

func (ps *PubSub) Unsubscribe(channelID string, ch chan TypingEvent) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, exists := ps.subscribers[channelID]; exists {
		delete(subs, ch)
		close(ch)
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
				// Skip if channel is full
			}
		}
	}
}
