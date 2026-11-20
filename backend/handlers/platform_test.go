package handlers

import "testing"

func TestTagImpersonation(t *testing.T) {
	actor := int64(7)

	got := tagImpersonation(nil, 5, &actor)
	if got["acting_as"] != int64(5) || got["actor"] != int64(7) {
		t.Fatalf("nil values: got %v", got)
	}

	in := map[string]interface{}{"name": "x"}
	got = tagImpersonation(in, 5, &actor)
	if got["name"] != "x" || got["acting_as"] != int64(5) {
		t.Fatalf("map values: got %v", got)
	}
	if _, mutated := in["acting_as"]; mutated {
		t.Fatal("input map must not be mutated")
	}

	got = tagImpersonation("raw", 5, nil)
	if got["data"] != "raw" {
		t.Fatalf("scalar values: got %v", got)
	}
	if _, ok := got["actor"]; ok {
		t.Fatal("actor must be omitted when unknown")
	}
}
