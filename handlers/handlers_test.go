package handlers

import (
	"testing"
	"time"
)

func TestMarkSeen_DedupsByID(t *testing.T) {
	seenMu.Lock()
	seenIDs = make(map[string]time.Time)
	seenMu.Unlock()

	if !markSeen("abc") {
		t.Fatal("first occurrence of an id should be processed")
	}
	if markSeen("abc") {
		t.Fatal("a repeated id should be skipped as a duplicate")
	}
	if !markSeen("def") {
		t.Fatal("a different id should be processed")
	}
	// Empty ids can't be deduped, so they must always be processed.
	if !markSeen("") || !markSeen("") {
		t.Fatal("empty id should always be processed")
	}
}
