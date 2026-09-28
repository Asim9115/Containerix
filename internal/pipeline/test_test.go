package pipeline

import (
	"testing"
	"time"

	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/types"
)

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

func TestStartContainer(t *testing.T) {
	deployment := repository.Deployment{
		ID:            "abcd",
		UserID:        "user-123",
		RepoURL:       "https://abc.com",
		Status:        types.DeployStopped,
		ContainerID:   "container-123",
		ImageTag:      "test-image:latest",
		HostPort:      10001,
		ContainerPort: 10000,
		TierName:      "Tier1",
		TierCPU:       0.2,
		TierMemory:    "500M",
		EnvJSON:       `{}`,
		Error:         "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
		h := &State{
		Repo: &repository.Repos{},
	}

	cfg, err := h.StartContainer(deployment)

	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("StartContainer() returned nil config")
	}

	if cfg.Tier.Name != "Tier1" {
		t.Errorf("expected Tier1, got %s", cfg.Tier.Name)
	}

	if cfg.Tier.Cpu != 0.2 {
		t.Errorf("expected CPU 0.2, got %v", cfg.Tier.Cpu)
	}

	if cfg.Tier.Memory != "500M" {
		t.Errorf("expected memory 500M, got %s", cfg.Tier.Memory)
	}
}