package types

// Deployment statuses
const (
    DeployQueued   = "queued"
    DeployBuilding = "building"
    DeployRunning  = "running"
    DeployStopped  = "stopped"
    DeployFailed   = "failed"
)

// Job statuses
const (
    JobQueued    = "queued"
    JobBuilding  = "building"
    JobCompleted = "completed"
    JobFailed    = "failed"
)