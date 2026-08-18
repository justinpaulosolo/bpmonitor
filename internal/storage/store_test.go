package storage

import "testing"

func TestOpen(t *testing.T) {
	store, err := Open(":memory")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if store == nil {
		t.Fatal("Open returned nil store")
	}

	defer store.Close()

	var name string
	err = store.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='readings'`).Scan(&name)
	if err != nil {
		t.Fatalf("readings table not found: %v", err)
	}
	if name != "readings" {
		t.Errorf("got table name %q, want %q", name, "readings")
	}
}
