package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements Database and is considered a
// provider. It implements all the interface methods
// which actually just return the *Queries methods.
type Postgres struct {
	Pool 	*pgxpool.Pool
	Repo 	*Queries
	Error 	error
}

// NewPostgres starts the connection with the database.
// It is used to return the *Conn.
func NewPostgres(c context.Context, connString string) *Postgres{
	pool, err := pgxpool.New(c, connString)
	if err != nil {
		return &Postgres{
			Error: err,
		}
	}
	return &Postgres{
		Pool: pool,
		Repo: New(pool),
	}
}

func (q *Postgres) GetError() error {
	return q.Error
}

// Close is used so that we can defer Database.Close.
// It return the pgx.Conn.Close method.
func (q *Postgres) Close(c context.Context) error{
	q.Pool.Close()
	return nil
}

func (q *Postgres) DeleteFile(c context.Context, objkey string) error {
	return q.Repo.DeleteFile(c, objkey)
}

func (q *Postgres) PutFile(c context.Context, arg PutFileParams) error {
	return q.Repo.PutFile(c, arg)
}

func (q *Postgres) GetUserFiles(c context.Context, ownerid string) ([]File,error) {
	return q.Repo.GetUserFiles(c, ownerid)
}

func (q *Postgres) GetAllFiles(c context.Context) ([]File,error) {
	return q.Repo.GetAllFiles(c)
}

func (q *Postgres) GetAllUsers(c context.Context) ([]User, error) {
	return q.Repo.GetAllUsers(c)
}

func (q *Postgres) GetUserSpace(ctx context.Context, username string) (int64, error) {
	return q.Repo.GetUserSpace(ctx, username)
}

func (q *Postgres) RecalculateUserSpace(ctx context.Context, username string) error {
	return q.Repo.RecalculateUserSpace(ctx, username)
}

func (q *Postgres) PutUser(ctx context.Context, username string) error {
	return q.Repo.PutUser(ctx, username)
}
