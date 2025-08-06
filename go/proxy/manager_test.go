package proxy

import "testing"

func TestManagerAddAndRetry(t *testing.T) {
	results := map[string]bool{
		"good": true,
		"bad":  false,
	}
	checker := func(addr string) bool { return results[addr] }

	m := NewManager(checker)
	m.Add("good")
	m.Add("bad")

	if len(m.Active()) != 1 {
		t.Fatalf("expected 1 active proxy, got %d", len(m.Active()))
	}
	if len(m.Failed()) != 1 {
		t.Fatalf("expected 1 failed proxy, got %d", len(m.Failed()))
	}

	results["bad"] = true
	if !m.Retry("bad") {
		t.Fatalf("expected retry to succeed")
	}
	if len(m.Failed()) != 0 {
		t.Fatalf("expected 0 failed proxies, got %d", len(m.Failed()))
	}
}

func TestDeleteAllFailed(t *testing.T) {
	results := map[string]bool{
		"bad1": false,
		"bad2": false,
	}
	checker := func(addr string) bool { return results[addr] }

	m := NewManager(checker)
	m.Add("bad1")
	m.Add("bad2")
	if len(m.Failed()) != 2 {
		t.Fatalf("expected 2 failed proxies, got %d", len(m.Failed()))
	}

	m.Delete("bad1")
	if len(m.Failed()) != 1 {
		t.Fatalf("expected 1 failed proxy, got %d", len(m.Failed()))
	}

	m.DeleteAllFailed()
	if len(m.Failed()) != 0 {
		t.Fatalf("expected no failed proxies after delete all")
	}
}
