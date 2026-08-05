package db2_storage

import (
	"context"
	"database/sql"
	"strings"

	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
	"github.com/storage-lock/go-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

// DB2Storage 基于 IBM DB2 数据库的存储引擎实现
// DB2 通过主键唯一约束和 WHERE 条件更新支持 CAS，
// 通过 CURRENT TIMESTAMP 提供数据库服务器时间
type DB2Storage struct {
	*sql_based_storage.SqlBasedStorage
	options *DB2StorageOptions
}

var _ storage.Storage = &DB2Storage{}

func NewDB2Storage(ctx context.Context, options *DB2StorageOptions) (*DB2Storage, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}

	baseStorageOption := sql_based_storage.NewSqlBasedStorageOptions().
		SetConnectionManager(options.ConnectionManager).
		SetSqlProvider(NewDB2SqlProvider()).
		SetTableFullName(options.TableName)
	baseStorage, err := sql_based_storage.NewSqlBasedStorage(baseStorageOption)
	if err != nil {
		return nil, err
	}

	s := &DB2Storage{
		SqlBasedStorage: baseStorage,
		options:         options,
	}

	if err := s.Init(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

const StorageName = "db2-storage"

func (x *DB2Storage) GetName() string {
	return StorageName
}

// Capabilities 声明 DB2 存储支持的能力
// DB2 通过主键唯一约束和 WHERE 条件更新天然支持 CAS，
// 通过 CURRENT TIMESTAMP 提供数据库服务器时间
func (x *DB2Storage) Capabilities() []storage.StorageCapability {
	return []storage.StorageCapability{
		storage.CapabilityCAS,
		storage.CapabilityReliableTime,
	}
}

// CreateWithVersion DB2 唯一约束冲突时返回 ErrLockAlreadyExists
func (x *DB2Storage) CreateWithVersion(ctx context.Context, lockId string, version storage.Version, lockInformation *storage.LockInformation) error {
	err := x.SqlBasedStorage.CreateWithVersion(ctx, lockId, version, lockInformation)
	if err != nil {
		// DB2 唯一约束冲突错误码: SQLSTATE=23505
		// 错误格式: SQLState: 23505, ErrorCode: -803
		msg := err.Error()
		if strings.Contains(msg, "23505") || strings.Contains(msg, "-803") {
			return storage_lock.ErrLockAlreadyExists
		}
	}
	return err
}

// DB2StorageOptions DB2 存储选项
type DB2StorageOptions struct {
	ConnectionManager storage.ConnectionManager[*sql.DB]
	TableName         string
}

func NewDB2StorageOptions() *DB2StorageOptions {
	return &DB2StorageOptions{
		TableName: storage.DefaultStorageTableName,
	}
}

func (x *DB2StorageOptions) SetConnectionManager(cm storage.ConnectionManager[*sql.DB]) *DB2StorageOptions {
	x.ConnectionManager = cm
	return x
}

func (x *DB2StorageOptions) SetTableName(name string) *DB2StorageOptions {
	x.TableName = name
	return x
}

func (x *DB2StorageOptions) Check() error {
	if x.ConnectionManager == nil {
		return ErrConnectionManagerNil
	}
	if x.TableName == "" {
		return ErrTableNameEmpty
	}
	return nil
}