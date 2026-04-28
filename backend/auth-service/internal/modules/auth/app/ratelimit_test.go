package app

import "testing"

func TestMergeRateLimit(t *testing.T) {
	t.Parallel()
	a := RateLimitInfo{Limit: 10, Remaining: 5, ResetUnix: 100}
	b := RateLimitInfo{Limit: 10, Remaining: 3, ResetUnix: 200}
	got := mergeRateLimit(a, b)
	if got.Remaining != 3 || got.ResetUnix != 200 || got.Limit != 10 {
		t.Fatalf("merge: %+v", got)
	}
	c := RateLimitInfo{Limit: 10, Remaining: 9, ResetUnix: 50}
	got = mergeRateLimit(a, c)
	if got.Remaining != 5 || got.ResetUnix != 100 {
		t.Fatalf("merge2: %+v", got)
	}
}
