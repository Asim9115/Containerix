package pipeline

import "testing"

func TestAgentTriage(t *testing.T) {
	result, err := AgentTriage(
		"test-job-123",
		"undefined: foo",
		"/tmp/test-repo",
		"https://github.com/example/test-repo",
		"go build ./... undefined: foo",
		
	)

	if err != nil {
		t.Fatalf("AgentTriage failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected response, got nil")
	}

	t.Logf("Job ID: %s", result.JobId)
	t.Logf("Root Cause: %s", result.RootCause)
	t.Logf("Fix: %s", result.Fix)
	t.Logf("Confidence: %s", result.Confidence)
	t.Logf("Steps Taken: %v", result.StepsTaken)
	t.Logf("Turns Used: %d", result.TurnsUsed)
}

