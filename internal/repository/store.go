package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound          = errors.New("resource not found")
	ErrorDuplicateEmail    = errors.New("this email has already been registerd")
	ErrorDuplicateUsername = errors.New("this username has already been registerd")
	QueryTimeOutDuration   = time.Second * 5
)

type Storage struct {
	Posts interface {
		GetById(context.Context, int64) (*Post, error)
		Create(context.Context, *Post) error
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFeedQuery) ([]PostWithMetadata, error)
	}

	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetUserById(context.Context, int64) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
		GetByEmail(context.Context, string) (*User, error)
	}
	Comment interface {
		Create(context.Context, *Comment) error
		GetByPostId(context.Context, int64) ([]Comment, error)
	}
	Followers interface {
		Follow(ctx context.Context, followerId int64, userId int64) error
		Unfollow(ctx context.Context, followerId int64, userId int64) error
	}
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostgresPostsStore{db},
		Users:     &PostgresUsersStore{db},
		Comment:   &PostgresCommentsStore{db},
		Followers: &FollowerRepository{db},
	}
}
func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
