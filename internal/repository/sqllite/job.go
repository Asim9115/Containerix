package sqllite

import (
	"database/sql"
	"time"

	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/types"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(job *repository.Job) error {
	createdAt := job.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	_, err := r.db.Exec(`
		INSERT INTO jobs (id, deployment_id, status, step, error, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, job.ID, job.DeploymentID, job.Status, job.Step, job.Error, createdAt)
	return err
}

func (r *JobRepo) GetByID(id string) (*repository.Job, error) {
	job := &repository.Job{}
	err := r.db.QueryRow(`
		SELECT id, deployment_id, status, COALESCE(step, ''), COALESCE(error, ''), created_at, completed_at 
		FROM jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.DeploymentID, &job.Status, &job.Step, &job.Error, &job.CreatedAt, &job.CompletedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (r *JobRepo) UpdateStatus(id, status, step string) error {
	_, err := r.db.Exec("UPDATE jobs SET status=?, step=? WHERE id=?", status, step, id)
	return err
}

func (r *JobRepo) SetFailed(id, errMsg string) error {
	_, err := r.db.Exec(`
		UPDATE jobs 
		SET status=?, error=?, completed_at=? 
		WHERE id=?
	`, types.JobFailed, errMsg, time.Now(), id)
	return err
}

func (r *JobRepo) SetCompleted(id, containerID string, hostPort int) error {
	_, err := r.db.Exec(`
		UPDATE jobs 
		SET status=?, completed_at=? 
		WHERE id=?
	`, types.JobCompleted, time.Now(), id)
	return err
}

func (r *JobRepo) GetAll() ([]repository.Job, error) {
	jobs := make([]repository.Job, 0)
	rows, err := r.db.Query(`
		SELECT id, deployment_id, status, COALESCE(step, ''), COALESCE(error, ''), created_at, completed_at 
		FROM jobs
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var job repository.Job
		if err := rows.Scan(&job.ID, &job.DeploymentID, &job.Status, &job.Step, &job.Error, &job.CreatedAt, &job.CompletedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *JobRepo) DeleteByDeploymentID(deploymentID string) error {
	_, err := r.db.Exec(`DELETE FROM jobs WHERE deployment_id = ? OR id = ?`, deploymentID, deploymentID)
	return err
}

func (r *JobRepo) ListByUser(userID string) ([]repository.Job, error) {
	var jobs []repository.Job
	rows, err := r.db.Query(`SELECT j.id, j.deployment_id, j.status, COALESCE(j.step, ''),
	COALESCE(j.error, ''), j.created_at, j.completed_at FROM jobs j
	JOIN deployments d ON d.id = j.deployment_id
	WHERE d.user_id=? ORDER BY j.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var job repository.Job
		if err = rows.Scan(&job.ID, &job.DeploymentID, &job.Status, &job.Step, &job.Error, &job.CreatedAt, &job.CompletedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}
