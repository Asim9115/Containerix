package sqllite

import (
	"database/sql"
	"time"

	"github.com/asim9115/containerix/internal/repository"
)

type PortsRepo struct {
	db *sql.DB
}

func NewPortsRepo(db *sql.DB) *PortsRepo {
	return &PortsRepo{db: db}
}


func (r *PortsRepo) FreePort(hostPort int) error {
	_, err := r.db.Exec(`DELETE FROM port_allocations WHERE host_port=?`, hostPort)
	return err
}

func (r *PortsRepo) Create(p *repository.Ports) error {
	_, err := r.db.Exec(`INSERT INTO port_allocations (host_port, container_id, container_port, allocated_at)
	VALUES (?, ?, ?, ?)`, p.HostPort, p.ContainerID, p.ContainerPort, time.Now())
	return err
}

func (r *PortsRepo) DeleteByContainerID(containerID string) error {
	_, err := r.db.Exec(`DELETE FROM port_allocations WHERE container_id=?`, containerID)
	return err
}

func (r *PortsRepo) GetAll() ([]repository.Ports, error) {
	var ports []repository.Ports
	rows, err := r.db.Query(`SELECT host_port, container_port, container_id, allocated_at FROM port_allocations`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var port repository.Ports

		err := rows.Scan(
			&port.HostPort,
			&port.ContainerPort,
			&port.ContainerID,
			&port.AllocatedAt,
		)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ports, nil
}