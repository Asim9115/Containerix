package sqllite

import (
	"database/sql"
	"time"

	"github.com/asim9115/containerix/internal/repository"
)

type DeploymentRepo struct {
	db *sql.DB
}

func NewDeploymentRepo(db *sql.DB) *DeploymentRepo {
	return &DeploymentRepo{db: db}
}

func (r *DeploymentRepo) Create(d *repository.Deployment) error {
	_, err := r.db.Exec(
		`INSERT INTO deployments
		(id, user_id, repo_url, status, tier_name, tier_cpu, tier_memory, env_json)
		VALUES (?,?,?,?,?,?,?,?)`, d.ID, d.UserID, d.RepoURL, d.Status, d.TierName, d.TierCPU, d.TierMemory, d.EnvJSON,
	)
	return err
}

func (r *DeploymentRepo) GetByID(id string) (*repository.Deployment, error) {
	d := &repository.Deployment{}
	err := r.db.QueryRow(
		`SELECT id, user_id, repo_url, status, container_id, image_tag,
                host_port, container_port, tier_name, error, created_at
         FROM deployments WHERE id = ?`, id,
	).Scan(&d.ID, &d.UserID, &d.RepoURL, &d.Status, &d.ContainerID,
		&d.ImageTag, &d.HostPort, &d.ContainerPort, &d.TierName,
		&d.Error, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (r *DeploymentRepo) UpdateStatus(id, status, ContainerID string, hostPort int) error {
	_, err := r.db.Exec(
		`UPDATE deployments SET status=?, container_id=?, host_port=?, updated_at=?
		WHERE id=?`, status, ContainerID, hostPort, time.Now(), id,
	)
	return err
}

func (r *DeploymentRepo) ListByUser(userID string) ([]repository.Deployment, error) {
	deployments := make([]repository.Deployment, 0)

	rows, err := r.db.Query(`
		SELECT
			id, user_id, repo_url, status,
			container_id,
			image_tag,
			host_port,
			container_port,
			tier_name,
			tier_cpu,
			tier_memory,
			env_json,
			error,
			created_at,
			updated_at
		FROM deployments
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var deployment repository.Deployment

		err := rows.Scan(
			&deployment.ID,
			&deployment.UserID,
			&deployment.RepoURL,
			&deployment.Status,
			&deployment.ContainerID,
			&deployment.ImageTag,
			&deployment.HostPort,
			&deployment.ContainerPort,
			&deployment.TierName,
			&deployment.TierCPU,
			&deployment.TierMemory,
			&deployment.EnvJSON,
			&deployment.Error,
			&deployment.CreatedAt,
			&deployment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, deployment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deployments, nil
}

func (r *DeploymentRepo) UpdateError(id, status, errMsg string) error {
	_, err := r.db.Exec(`
		UPDATE deployments
		SET error=?, status=?, updated_at=?
		WHERE id=?
	`, errMsg, status, time.Now(), id)
	
	return err
}

func (r *DeploymentRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM deployments where id=?`, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *DeploymentRepo) ListByStatus(status string) ([]repository.Deployment, error) {
	deployments := make([]repository.Deployment, 0)
	rows, err := r.db.Query(`SELECT id, user_id, repo_url, status,
			container_id,
			image_tag,
			host_port,
			container_port,
			tier_name,
			tier_cpu,
			tier_memory,
			env_json,
			error,
			created_at,
			updated_at FROM deployments
			WHERE status=?`, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var deployment repository.Deployment

		err := rows.Scan(
			&deployment.ID,
			&deployment.UserID,
			&deployment.RepoURL,
			&deployment.Status,
			&deployment.ContainerID,
			&deployment.ImageTag,
			&deployment.HostPort,
			&deployment.ContainerPort,
			&deployment.TierName,
			&deployment.TierCPU,
			&deployment.TierMemory,
			&deployment.EnvJSON,
			&deployment.Error,
			&deployment.CreatedAt,
			&deployment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, deployment)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deployments, nil
}
func (r *DeploymentRepo)GetAll() ([]repository.Deployment, error) {
	deployments := make([]repository.Deployment, 0)
	rows, err := r.db.Query(`SELECT id, user_id, repo_url, status,
			container_id,
			image_tag,
			host_port,
			container_port,
			tier_name,
			tier_cpu,
			tier_memory,
			env_json,
			error,
			created_at,
			updated_at FROM deployments
			`,)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var deployment repository.Deployment

		err := rows.Scan(
			&deployment.ID,
			&deployment.UserID,
			&deployment.RepoURL,
			&deployment.Status,
			&deployment.ContainerID,
			&deployment.ImageTag,
			&deployment.HostPort,
			&deployment.ContainerPort,
			&deployment.TierName,
			&deployment.TierCPU,
			&deployment.TierMemory,
			&deployment.EnvJSON,
			&deployment.Error,
			&deployment.CreatedAt,
			&deployment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, deployment)

	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deployments, nil
}