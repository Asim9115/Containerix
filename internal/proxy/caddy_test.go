package proxy

import "testing"

func TestAddRoute(t *testing.T) {
	err := AddRoute("localhost:4009", "localhost:5173")
	if err != nil {
		t.Fatalf("AddRoute failed: %v", err)
	}
}