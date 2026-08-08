package sqllite

import (
	"database/sql"
	"time"

	"github.com/asim9115/containerix/internal/repository"
)


type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(u *repository.User) error {
	_, err := r.db.Exec(`INSERT INTO users (id, name, email, api_key_hash, created_at) VALUES (?, ?, ?, ?, ?)`, u.ID, u.Name,
	u.Email, u.ApiKeyHash, time.Now())
	return err
}

func (r *UserRepo) Delete(ID string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id = ? `, ID)
	return err
}

func (r *UserRepo) UpdateEmail(ID string, email string) error {
	_, err := r.db.Exec(`UPDATE users SET email=? WHERE id=?`, email, ID)
	return err
}

func (r *UserRepo) UpdateName(ID string, name string) error {
	_, err := r.db.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, ID)
	return err
}

func (r *UserRepo) GetUser(ID string) (*repository.User, error) {
	user := &repository.User{}
	err := r.db.QueryRow(`SELECT id, name, email, api_key_hash, created_at FROM users WHERE id=?`, ID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.ApiKeyHash,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetAll() ([]repository.User, error) {
	users := make([]repository.User, 0)
	rows, err := r.db.Query(`SELECT id, name, email, api_key_hash, created_at from users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user repository.User
		err := rows.Scan(
			&user.ID,
			&user.Name, &user.Email,
			&user.ApiKeyHash,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) GetByApiKeyHash(key string) (*repository.User, error) {
	user := &repository.User{}
	err := r.db.QueryRow(`SELECT id, name, email, api_key_hash, created_at FROM users WHERE api_key_hash=?`, key).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.ApiKeyHash,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo)UpdateApiKeyHash(ID string, key string) error{
	_, err := r.db.Exec(`UPDATE users SET api_key_hash=? WHERE id=?`, key, ID)
	return err
}