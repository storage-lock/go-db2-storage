package db2_storage

import (
	"context"
	"database/sql"
	"sync"

	"github.com/storage-lock/go-storage"
)

// DB2ConnectionManager DB2 连接管理器
type DB2ConnectionManager struct {
	ConnectionString string

	mu sync.Mutex
	db *sql.DB

	open        func(driverName, dataSourceName string) (*sql.DB, error)
	ping        func(context.Context, *sql.DB) error
	initialized bool
}

var _ storage.ConnectionManager[*sql.DB] = &DB2ConnectionManager{}

// NewDB2ConnectionManagerFromDSN 从数据源名称创建连接管理器
// DB2 连接字符串格式: HOSTNAME=host;DATABASE=db;PORT=port;UID=username;PWD=password
func NewDB2ConnectionManagerFromDSN(dsn string) *DB2ConnectionManager {
	return &DB2ConnectionManager{
		ConnectionString: dsn,
	}
}

func (x *DB2ConnectionManager) Name() string {
	return "db2-connection-manager"
}

func (x *DB2ConnectionManager) Take(ctx context.Context) (*sql.DB, error) {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.initialized {
		return x.db, nil
	}

	// DB2 驱动名称为 "go_ibm_db"
	open := x.open
	if open == nil {
		open = sql.Open
	}
	db, err := open("go_ibm_db", x.ConnectionString)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 测试连接
	ping := x.ping
	if ping == nil {
		ping = func(ctx context.Context, db *sql.DB) error { return db.PingContext(ctx) }
	}
	if err := ping(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	x.db = db
	x.initialized = true
	return x.db, nil
}

func (x *DB2ConnectionManager) Return(ctx context.Context, conn *sql.DB) error {
	return nil
}

func (x *DB2ConnectionManager) Shutdown(ctx context.Context) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.db != nil {
		return x.db.Close()
	}
	return nil
}
