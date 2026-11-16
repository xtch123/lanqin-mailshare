package app

import (
	"reflect"
	"testing"

	"github.com/emersion/go-imap/v2"
)

func TestBoundedSearchUIDsAfterIgnoresStaleServerMatches(t *testing.T) {
	// A broken server may return all matching mailbox UIDs even when UID
	// SEARCH requested only UID > 483607. Filter before limiting to 10.
	var all imap.UIDSet
	all.AddRange(483400, 483609)
	got := boundedSearchUIDsAfter(all, 483607, 10)
	want := []imap.UID{483608, 483609}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got UIDs %v, want %v", got, want)
	}
	if got := boundedSearchUIDsAfter(all, 483609, 10); len(got) != 0 {
		t.Fatalf("stale messages must not be returned: %v", got)
	}
}

func TestBoundedSearchUIDsAfterHonorsLimitsAndUIDGaps(t *testing.T) {
	var all imap.UIDSet
	all.AddRange(10, 12)
	all.AddRange(20, 30)
	got := boundedSearchUIDsAfter(all, 11, 4)
	want := []imap.UID{12, 20, 21, 22}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got UIDs %v, want %v", got, want)
	}
}

func TestBoundedSearchUIDsAfterMaxUIDDoesNotWrap(t *testing.T) {
	var all imap.UIDSet
	all.AddRange(imap.UID(^uint32(0)-1), imap.UID(^uint32(0)))
	got := boundedSearchUIDsAfter(all, ^uint32(0)-2, 10)
	want := []imap.UID{imap.UID(^uint32(0)-1), imap.UID(^uint32(0))}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got UIDs %v, want %v", got, want)
	}
	if got := boundedSearchUIDsAfter(all, ^uint32(0), 10); len(got) != 0 {
		t.Fatalf("maximum UID must not wrap: %v", got)
	}
}
