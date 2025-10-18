package storage

import (
	"testing"
	"time"
)

func TestMemoryStore_SaveListGet(t *testing.T) {
	ms := NewMemoryStore(500 * time.Millisecond)
	defer ms.Close()

	addr := ms.CreateAddress("test")
	saved, err := ms.Save(addr, Message{From: "a@b", Subject: "hello", Snippet: "world"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == "" {
		t.Fatal("no id")
	}

	list := ms.List(addr)
	if len(list) != 1 {
		t.Fatalf("want 1, got %d", len(list))
	}

	got, ok := ms.Get(addr, saved.ID)
	if !ok {
		t.Fatal("not found")
	}
	if got.Subject != "hello" {
		t.Fatalf("bad subject: %s", got.Subject)
	}
}

func TestMemoryStore_TTL(t *testing.T) {
	ms := NewMemoryStore(100 * time.Millisecond)
	defer ms.Close()
	addr := ms.CreateAddress("ttl")
	_, _ = ms.Save(addr, Message{Subject: "x"})
	time.Sleep(150 * time.Millisecond)
	ms.PurgeExpired()
	if len(ms.List(addr)) != 0 {
		t.Fatal("expected expired")
	}
}
