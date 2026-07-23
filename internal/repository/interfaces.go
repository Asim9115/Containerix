package repository

import "time"

type Deployment struct {
    ID            string
    UserID        string
    RepoURL       string
    Status        string
    ContainerID   string
    ImageTag      string
    HostPort      int
    ContainerPort int
    TierName      string
    TierCPU       float64
    TierMemory    string
    EnvJSON       string
    Error         string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Job struct {
    ID           string
    DeploymentID string
    Status       string
    Step         string
    Error        string
    CreatedAt    time.Time
    CompletedAt  *time.Time
}

// DeploymentRepo — swap the implementation to switch databases
type DeploymentRepo interface {
    Create(d *Deployment) error
    GetByID(id string) (*Deployment, error)
    ListByUser(userID string) ([]Deployment, error)
    UpdateStatus(id, status, containerID string, hostPort int) error
    UpdateError(id, status, errMsg string) error
    Delete(id string) error
    ListByStatus(status string) ([]Deployment, error)
    GetAll() ([]Deployment, error)
}

// JobRepo — swap the implementation to switch databases
type JobRepo interface {
    Create(j *Job) error
    GetByID(id string) (*Job, error)
    UpdateStatus(id, status, step string) error
    SetFailed(id, errMsg string) error
    SetCompleted(id, containerID string, hostPort int) error
    GetAll() ([]Job, error)
}
