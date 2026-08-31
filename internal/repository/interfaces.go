package repository

import (
	"time"

)

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

type Ports struct {
    HostPort int
    ContainerID string
    ContainerPort int
    AllocatedAt time.Time
}

type User struct {
    ID string
    Name string
    Email string
    ApiKeyHash string
    CreatedAt time.Time
}

// DeploymentRepo — swap the implementation to switch databases
type DeploymentRepo interface {
    Create(d *Deployment) error
    GetByID(id string) (*Deployment, error)
    ListByUser(userID string) ([]Deployment, error)
    UpdateStatus(id, status, containerID, imageTag string, hostPort, containerPort int) error
    UpdateError(id, status, errMsg string) error
    Delete(id string) error
    DeleteByContainerID(containerID string) error
    ListByStatus(status string) ([]Deployment, error)
    GetAll() ([]Deployment, error)
    GetByContainerId(containerID string) (*Deployment, error)
    UpdateStatusByContainerID(containerID string, status string) error
    UpdateStatusAndPort(containerID string, status string, hostPort int) error
}

// JobRepo — swap the implementation to switch databases
type JobRepo interface {
    Create(j *Job) error
    GetByID(id string) (*Job, error)
    ListByUser(userID string) ([]Job, error)
    UpdateStatus(id, status, step string) error
    SetFailed(id, errMsg string) error
    SetCompleted(id, containerID string, hostPort int) error
    GetAll() ([]Job, error)
    DeleteByDeploymentID(deploymentID string) error
}

//PortsRepo
type PortsRepo interface {
    Create(p *Ports) error
    FreePort(HostPort int) error
    DeleteByContainerID(ContainerID string) error
    GetAll() ([]Ports, error)
}

//User
type UserRepo interface {
    Create(u *User) error
    Delete(ID string) error
    UpdateEmail(ID string, email string) error
    UpdateName(ID string, name string) error
    UpdateApiKeyHash(ID string, key string) error
    GetUser(ID string) (*User, error)
    GetByApiKeyHash(key string) (*User, error)
    GetAll() ([]User, error)
}

type Repos struct {
    Deployments DeploymentRepo
    Jobs JobRepo
    Ports PortsRepo
    User UserRepo
}