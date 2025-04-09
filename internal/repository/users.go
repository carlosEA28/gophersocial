package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash
	return nil
}

type PostgresUsersStore struct {
	db *sql.DB
}

func (s *PostgresUsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `INSERT INTO users (username, password, email) VALUES($1,$2,$3) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	err := tx.QueryRowContext(ctx, query, user.Username, user.Password.hash, user.Email).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrorDuplicateEmail

		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrorDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *PostgresUsersStore) GetUserById(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id,username,email,password,created_at FROM users WHERE id = $1 AND is_active = true`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *PostgresUsersStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {

		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err // Retorna o erro corretamente
		}

		return nil
	})
}

func (s *PostgresUsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userId int64) error {
	query := `INSERT INTO user_invitations (token,user_id,expiry) VALUES ($1,$2,$3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userId, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresUsersStore) getUserByInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `SELECT u.id,u.username,u.email,u.created_at,u.is_active FROM users u JOIN user_invitations ui ON u.id = ui.user_id WHERE ui.token = $1 AND ui.expiry > $2`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *PostgresUsersStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `UPDATE users SET username=$1, email=$2, is_active=$3 WHERE id=$4
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresUsersStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresUsersStore) Activate(ctx context.Context, token string) error {

	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		//find the user by the token
		user, err := s.getUserByInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		//update the user
		user.IsActive = true
		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		//delete the invitations
		if err := s.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil

	})
}

func (s *PostgresUsersStore) Delete(ctx context.Context, userId int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userId); err != nil {
			return err
		}
		if err := s.deleteUserInvitations(ctx, tx, userId); err != nil {
			return err
		}

		return nil
	})
}

func (s *PostgresUsersStore) delete(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresUsersStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id,username,email,password,created_at FROM users
	WHERE email = $1 AND is_active = true`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	user := &User{}

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
