package config

import (
	"errors"
	"testing"
)

func TestStore_GetReturnsCurrentConfig(t *testing.T) {
	cfg := &Config{ServerURL: "http://original"}
	saveFn := func(*Config) error { return nil }
	store := NewStore(cfg, saveFn)

	got := store.Get()
	if got.ServerURL != "http://original" {
		t.Errorf("expected http://original, got %s", got.ServerURL)
	}
}

func TestStore_UpdatePersists(t *testing.T) {
	cfg := &Config{ServerURL: "http://old"}
	saveCalls := 0
	saveFn := func(c *Config) error {
		saveCalls++
		if c.ServerURL != "http://new" {
			t.Errorf("save received stale config: %s", c.ServerURL)
		}
		return nil
	}
	store := NewStore(cfg, saveFn)

	err := store.Update(func(c *Config) {
		c.ServerURL = "http://new"
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saveCalls != 1 {
		t.Errorf("expected 1 save call, got %d", saveCalls)
	}
	if store.Get().ServerURL != "http://new" {
		t.Errorf("in-memory config not updated")
	}
}

func TestStore_UpdatePropagatesSaveError(t *testing.T) {
	cfg := &Config{}
	sentinel := errors.New("disk full")
	store := NewStore(cfg, func(*Config) error { return sentinel })

	err := store.Update(func(c *Config) {
		c.ServerURL = "http://x"
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	// Mutation still applied even if save failed — the store reflects the attempt
	if store.Get().ServerURL != "http://x" {
		t.Errorf("mutation should still apply on save failure")
	}
}

func TestStore_UpdateNoSaveDoesNotPersist(t *testing.T) {
	cfg := &Config{}
	saveCalls := 0
	store := NewStore(cfg, func(*Config) error {
		saveCalls++
		return nil
	})

	store.UpdateNoSave(func(c *Config) {
		c.ServerURL = "http://ephemeral"
	})

	if saveCalls != 0 {
		t.Errorf("UpdateNoSave should not call save, got %d calls", saveCalls)
	}
	if store.Get().ServerURL != "http://ephemeral" {
		t.Errorf("mutation should still apply")
	}
}

func TestStore_DefaultSaveFnFallback(t *testing.T) {
	// Passing nil saveFn should fall back to config.Save (not tested here,
	// just verify the store is constructable).
	cfg := &Config{}
	store := NewStore(cfg, nil)
	if store == nil {
		t.Error("NewStore with nil saveFn should not return nil")
	}
}
