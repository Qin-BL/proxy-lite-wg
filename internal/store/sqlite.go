package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"

	"proxy-lite-wg/internal/domain"
)

var ErrNotFound = errors.New("not found")

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) Migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS clients (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			label TEXT NOT NULL,
			state TEXT NOT NULL,
			client_uuid TEXT NOT NULL UNIQUE,
			disabled_at TIMESTAMP NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_clients_user_id ON clients(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_clients_state ON clients(state);`,
	}

	for _, statement := range statements {
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteRepository) CreateUser(ctx context.Context, user domain.User) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, name, email, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		user.ID,
		user.Name,
		user.Email,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, email, status, created_at, updated_at
		 FROM users
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *SQLiteRepository) GetUserByID(ctx context.Context, id string) (domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, name, email, status, created_at, updated_at
		 FROM users WHERE id = ?`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, ErrNotFound
	}
	return user, err
}

func (r *SQLiteRepository) DeleteUser(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLiteRepository) CountClientsByUser(ctx context.Context, userID string) (int, error) {
	var count int
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(1) FROM clients WHERE user_id = ?`,
		userID,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SQLiteRepository) SaveClient(ctx context.Context, client domain.Client) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO clients (
			id, user_id, label, state, client_uuid, disabled_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			user_id = excluded.user_id,
			label = excluded.label,
			state = excluded.state,
			client_uuid = excluded.client_uuid,
			disabled_at = excluded.disabled_at,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at`,
		client.ID,
		client.UserID,
		client.Label,
		client.State,
		client.ClientUUID,
		client.DisabledAt,
		client.CreatedAt,
		client.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepository) ListClients(ctx context.Context) ([]domain.Client, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, label, state, client_uuid, disabled_at, created_at, updated_at
		 FROM clients
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []domain.Client
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, rows.Err()
}

func (r *SQLiteRepository) ListActiveClients(ctx context.Context) ([]domain.Client, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, user_id, label, state, client_uuid, disabled_at, created_at, updated_at
		 FROM clients
		 WHERE state = 'active'
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []domain.Client
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, rows.Err()
}

func (r *SQLiteRepository) GetClientByID(ctx context.Context, id string) (domain.Client, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, label, state, client_uuid, disabled_at, created_at, updated_at
		 FROM clients WHERE id = ?`,
		id,
	)

	client, err := scanClient(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Client{}, ErrNotFound
	}
	return client, err
}

func (r *SQLiteRepository) DeleteClient(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM clients WHERE id = ?`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanClient(s scanner) (domain.Client, error) {
	var client domain.Client
	var disabledAt sql.NullTime
	err := s.Scan(
		&client.ID,
		&client.UserID,
		&client.Label,
		&client.State,
		&client.ClientUUID,
		&disabledAt,
		&client.CreatedAt,
		&client.UpdatedAt,
	)
	if err != nil {
		return domain.Client{}, err
	}
	if disabledAt.Valid {
		client.DisabledAt = &disabledAt.Time
	}
	return client, nil
}
