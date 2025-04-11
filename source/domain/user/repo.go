package user

import (
	"context"
	"database/sql"
	"time"
)

type SQLRepo struct {
	Client *sql.DB
}

// createUser inserts a new user into the database.
func (repo *SQLRepo) Insert(ctx context.Context, user *User) error {
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return repo.Client.QueryRow(query, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// getUserByID fetches a user by their ID from the database.
func (repo *SQLRepo) Find(ctx context.Context, id uint64) (*User, error) {
	user := User{}
	query := `SELECT * FROM users WHERE id = $1`
	err := repo.Client.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

// updateUser modifies an existing user record in the database.
func (repo *SQLRepo) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now() // Set the updated time to now
	query := `UPDATE users SET email = $1, password_hash = $2, updated_at = $3 WHERE id = $4`
	_, err := repo.Client.Exec(query, user.Email, user.PasswordHash, user.UpdatedAt, user.ID)
	return err
}

// deleteUser removes a user record from the database.
func (repo *SQLRepo) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := repo.Client.Exec(query, id)
	return err
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Users  []User
	Cursor uint64
}

func (repo *SQLRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	users := []User{}
	query := `SELECT * FROM users ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := repo.Client.Query(query, page.Size, page.Offset)
	if err != nil {
		return FindResult{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return FindResult{}, err
		}
		users = append(users, user)
	}

	defer rows.Close()

	// The cursor could be set to the next offset or a more complex pagination token
	nextCursor := page.Offset + uint64(len(users))
	return FindResult{Users: users, Cursor: nextCursor}, nil
}
