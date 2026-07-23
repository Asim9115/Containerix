package sqllite

import (
	"github.com/asim9115/containerix/internal/repository"
		"database/sql"
)

// type JobRepo interface {
//     Create(j *Job) error
//     GetByID(id string) (*Job, error)
//     UpdateStatus(id, status, step string) error
//     SetFailed(id, errMsg string) error
//     SetCompleted(id, containerID string, hostPort int) error
//     GetAll() ([]Job, error)
// }
// CREATE TABLE IF NOT EXISTS jobs (
//     id TEXT PRIMARY KEY,
//     deployment_id TEXT,
//     status TEXT NOT NULL DEFAULT 'queued',
//     step TEXT,
//     error TEXT,
//     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//     completed_at DATETIME,
//     FOREIGN KEY (deployment_id) REFERENCES deployments(id)
// );
type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(job repository.Job) error {
	_, err := r.db.Query(`INSERT INTO job (id, deployment_id, status, step, error, created_at, )`)
}