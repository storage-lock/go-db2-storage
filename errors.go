package db2_storage

import "errors"

var (
	ErrConnectionManagerNil = errors.New("ConnectionManager cannot be nil")
	ErrTableNameEmpty       = errors.New("TableName cannot be empty")
)