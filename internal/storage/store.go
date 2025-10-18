package storage

import (
    "sort"
    "sync"
    "time"

    "github.com/google/uuid"
)

type Message struct {
    ID        string    `json:"id"`
    Address   string    `json:"address"`
    From      string    `json:"from"`
    Subject   string    `json:"subject"`
    Snippet   string    `json:"snippet"`
    CreatedAt time.Time `json:"createdAt"`
    ExpiresAt time.Time `json:"expiresAt"`
    // Raw MIME for full fetch
    Raw []byte `json:"-"`
}

type Store interface {
    CreateAddress(local string) string
    AddressExists(local string) bool
    Save(addr string, msg Message) (Message, error)
    List(addr string) []Message
    Get(addr, id string) (Message, bool)
    PurgeExpired()
    TTL() time.Duration
    Close()
}

type MemoryStore struct {
    mu       sync.RWMutex
    ttl      time.Duration
    messages map[string]map[string]Message // addr -> id -> message
    stopCh   chan struct{}
}

func NewMemoryStore(ttl time.Duration) *MemoryStore {
    ms := &MemoryStore{
        ttl:      ttl,
        messages: make(map[string]map[string]Message),
        stopCh:   make(chan struct{}),
    }
    go ms.gcLoop()
    return ms
}

func (m *MemoryStore) CreateAddress(local string) string {
    if local == "" {
        local = uuid.NewString()
    }
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.messages[local]; !ok {
        m.messages[local] = make(map[string]Message)
    }
    return local
}

func (m *MemoryStore) AddressExists(local string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    _, exists := m.messages[local]
    return exists
}

func (m *MemoryStore) Save(addr string, msg Message) (Message, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.messages[addr]; !ok {
        m.messages[addr] = make(map[string]Message)
    }
    if msg.ID == "" {
        msg.ID = uuid.NewString()
    }
    now := time.Now()
    msg.CreatedAt = now
    msg.ExpiresAt = now.Add(m.ttl)
    m.messages[addr][msg.ID] = msg
    return msg, nil
}

func (m *MemoryStore) List(addr string) []Message {
    m.mu.RLock()
    defer m.mu.RUnlock()
    var out []Message
    for _, msg := range m.messages[addr] {
        out = append(out, msg)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
    return out
}

func (m *MemoryStore) Get(addr, id string) (Message, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    msgs, ok := m.messages[addr]
    if !ok {
        return Message{}, false
    }
    msg, ok := msgs[id]
    if !ok {
        return Message{}, false
    }
    return msg, true
}

func (m *MemoryStore) PurgeExpired() {
    m.mu.Lock()
    defer m.mu.Unlock()
    now := time.Now()
    for addr, msgs := range m.messages {
        for id, msg := range msgs {
            if now.After(msg.ExpiresAt) {
                delete(msgs, id)
            }
        }
        if len(msgs) == 0 {
            delete(m.messages, addr)
        }
    }
}

func (m *MemoryStore) TTL() time.Duration { return m.ttl }

func (m *MemoryStore) gcLoop() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            m.PurgeExpired()
        case <-m.stopCh:
            return
        }
    }
}

func (m *MemoryStore) Close() {
    close(m.stopCh)
}
