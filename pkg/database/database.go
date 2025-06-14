package database

import "context"

// Database interface implement some methods that 
// every provider should follow.
// Those methods are used for managing file and user metadata
// not the files themselves. 
// For that take a look at package storage
type Database interface {
	GetAllFiles(context.Context) ([]File, error)
	GetUserFiles(context.Context, string) ([]File, error)
	PutFile(context.Context, PutFileParams) error
	DeleteFile(context.Context, string) error

	GetError() error 

	GetAllUsers(context.Context) ([]User, error)
	PutUser(context.Context, string) error
	GetUserSpace(context.Context, string) (int64, error)
	RecalculateUserSpace(context.Context, string) error

	Close(context.Context) error
}
