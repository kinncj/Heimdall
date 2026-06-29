// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>

package discovery

import "testing"

func TestSortedFoundDedupesAndSorts(t *testing.T) {
	got := sortedFound(map[string]Found{
		"b:1": {Name: "b", Addr: "b:1"},
		"a:2": {Name: "a", Addr: "a:2"},
		"a:1": {Name: "a", Addr: "a:1"},
	})
	if len(got) != 3 {
		t.Fatalf("got %d, want 3", len(got))
	}
	// sorted by name then addr
	want := []Found{{Name: "a", Addr: "a:1"}, {Name: "a", Addr: "a:2"}, {Name: "b", Addr: "b:1"}}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pos %d = %v, want %v", i, got[i], want[i])
		}
	}
	if len(sortedFound(map[string]Found{})) != 0 {
		t.Fatal("empty map should yield an empty slice")
	}
}
