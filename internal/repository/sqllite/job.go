package sqllite

import (
	"database/sql"
	"time"

	"github.com/asim9115/containerix/internal/repository"
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
		SET status='failed', error=?, completed_at=? 
		WHERE id=?
	`, errMsg, time.Now(), id)
	return err
}

func (r *JobRepo) SetCompleted(id, containerID string, hostPort int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update the job status
	_, err = tx.Exec(`
		UPDATE jobs 
		SET status='completed', completed_at=? 
		WHERE id=?
	`, time.Now(), id)
	if err != nil {
		return err
	}

	// 2. Find the deployment ID associated with the job
	var deploymentID string
	err = tx.QueryRow(`SELECT deployment_id FROM jobs WHERE id=?`, id).Scan(&deploymentID)
	if err != nil {
		return err
	}

	// 3. Update the deployment status to active and record container details
	if deploymentID != "" {
		_, err = tx.Exec(`
			UPDATE deployments 
			SET status='active', container_id=?, host_port=?, updated_at=? 
			WHERE id=?
		`, containerID, hostPort, time.Now(), deploymentID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
