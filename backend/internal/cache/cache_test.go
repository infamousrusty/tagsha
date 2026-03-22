// SPDX-License-Identifier: LGPL-3.0-or-later
// Copyright (C) 2026 infamousrusty

package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/infamousrusty/tagsha/internal/cache"
)

// newTestCache spins up an in-process Redis stub and returns a Cache backed by it.
func newTestCache(t *testing.T) (*cache.Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	c, err := cache.New("redis://"+mr.Addr(), 5*time.Minute)
	if err != nil {
		t.Fatalf("cache.New: %v", err)
	}
	return c, mr
}

type payload struct {
	Owner string `json:"owner"`
	Count int    `json:"count"`
}

// ---------------------------------------------------------------------------
// TagsKey
// ---------------------------------------------------------------------------

func TestTagsKey(t *testing.T) {
	key := cache.TagsKey("golang", "go")
	if key != "tagsha:tags:golang:go" {
		t.Errorf("TagsKey = %q, want tagsha:tags:golang:go", key)
	}
}

// ---------------------------------------------------------------------------
// Get / Set — primary cache
// ---------------------------------------------------------------------------

func TestSetAndGet_Hit(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	want := payload{Owner: "golang", Count: 42}
	if err := c.Set(ctx, "testkey", want, 0); err != nil {
		t.Fatalf("Set: %v", err)
	}

	var got payload
	hit, err := c.Get(ctx, "testkey", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !hit {
		t.Fatal("expected cache HIT, got MISS")
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGet_Miss(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	var got payload
	hit, err := c.Get(ctx, "nonexistent", &got)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if hit {
		t.Fatal("expected cache MISS, got HIT")
	}
}

// ---------------------------------------------------------------------------
// TTL expiry
// ---------------------------------------------------------------------------

func TestSet_TTLExpiry(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	want := payload{Owner: "expired", Count: 1}
	if err := c.Set(ctx, "ttlkey", want, 1*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Advance miniredis clock past TTL.
	mr.FastForward(2 * time.Second)

	var got payload
	hit, err := c.Get(ctx, "ttlkey", &got)
	if err != nil {
		t.Fatalf("Get after expiry: %v", err)
	}
	if hit {
		t.Fatal("expected MISS after TTL expiry, got HIT")
	}
}

// ---------------------------------------------------------------------------
// Default TTL is applied when ttl == 0
// ---------------------------------------------------------------------------

func TestSet_DefaultTTL(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	// default TTL is 5 min. Advance 4 min — should still be present.
	if err := c.Set(ctx, "defttl", payload{Owner: "x", Count: 1}, 0); err != nil {
		t.Fatalf("Set: %v", err)
	}
	mr.FastForward(4 * time.Minute)

	var got payload
	hit, _ := c.Get(ctx, "defttl", &got)
	if !hit {
		t.Error("expected HIT within default TTL window")
	}

	// Advance past default TTL — should now be MISS.
	mr.FastForward(2 * time.Minute)
	hit, _ = c.Get(ctx, "defttl", &got)
	if hit {
		t.Error("expected MISS after default TTL expiry")
	}
}

// ---------------------------------------------------------------------------
// Stale fallback (SetStale / GetStale)
// ---------------------------------------------------------------------------

func TestStale_SetAndGet(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	want := payload{Owner: "staleOwner", Count: 99}
	if err := c.SetStale(ctx, "stalekey", want); err != nil {
		t.Fatalf("SetStale: %v", err)
	}

	var got payload
	hit, err := c.GetStale(ctx, "stalekey", &got)
	if err != nil {
		t.Fatalf("GetStale: %v", err)
	}
	if !hit {
		t.Fatal("expected stale HIT")
	}
	if got != want {
		t.Errorf("stale got %+v, want %+v", got, want)
	}
}

func TestStale_IndependentOfPrimary(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	v := payload{Owner: "x", Count: 1}
	_ = c.Set(ctx, "key", v, 1*time.Second)
	_ = c.SetStale(ctx, "key", v)

	// Primary expires, stale should survive.
	mr.FastForward(2 * time.Second)

	var got payload
	primaryHit, _ := c.Get(ctx, "key", &got)
	if primaryHit {
		t.Error("primary should be expired")
	}

	staleHit, _ := c.GetStale(ctx, "key", &got)
	if !staleHit {
		t.Error("stale fallback should still be present")
	}
}

// ---------------------------------------------------------------------------
// Unmarshal error — corrupted cache entry
// ---------------------------------------------------------------------------

func TestGet_CorruptEntry(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	// Write invalid JSON directly to Redis.
	mr.Set("corrupt", "not-json")

	var got payload
	hit, err := c.Get(ctx, "corrupt", &got)
	if err == nil {
		t.Error("expected unmarshal error for corrupt entry, got nil")
	}
	if hit {
		t.Error("hit should be false on unmarshal error")
	}
	_ = json.Unmarshal // ensure import used
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

func TestPing_Healthy(t *testing.T) {
	c, _ := newTestCache(t)
	if err := c.Ping(context.Background()); err != nil {
		t.Errorf("Ping: %v", err)
	}
}

func TestPing_Down(t *testing.T) {
	c, mr := newTestCache(t)
	// Stop miniredis to simulate Redis unavailability.
	mr.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := c.Ping(ctx); err == nil {
		t.Error("expected Ping error when Redis is down, got nil")
	}
}
