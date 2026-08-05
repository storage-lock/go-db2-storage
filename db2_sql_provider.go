package db2_storage

import (
	"context"
	sql_based_storage "github.com/storage-lock/go-sql-based-storage"
)

// DB2SqlProvider IBM DB2 数据库的 SQL 方言提供者
// DB2 使用 SQL PL 方言，与标准 SQL 有一些差异：
//   - 时间函数：CURRENT_TIMESTAMP（而非 NOW()）
//   - 字符串连接：使用 || 而非 CONCAT()
//   - 限制行数：使用 FETCH FIRST n ROWS ONLY 而非 LIMIT
type DB2SqlProvider struct {
	*sql_based_storage.Sql92Provider
}

func NewDB2SqlProvider() *DB2SqlProvider {
	return &DB2SqlProvider{
		Sql92Provider: sql_based_storage.NewSql92Provider(),
	}
}

// CreateTableSql DB2 创建表的 SQL
func (x *DB2SqlProvider) CreateTableSql(ctx context.Context, tableFullName string) (string, []any) {
	sql := `CREATE TABLE ` + tableFullName + ` (
		lock_id VARCHAR(255) PRIMARY KEY NOT NULL,
		version BIGINT NOT NULL,
		json_data CLOB(16M) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT TIMESTAMP
	)`
	return sql, nil
}

// NowTimestampSql DB2 获取当前时间戳
// DB2 使用 (DAYS(CURRENT DATE) - DAYS('1970-01-01')) * 86400 +
//           MIDNIGHT_SECONDS(CURRENT TIMESTAMP) 获取 Unix 时间戳
func (x *DB2SqlProvider) NowTimestampSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT (DAYS(CURRENT DATE) - DAYS('1970-01-01')) * 86400 + MIDNIGHT_SECONDS(CURRENT TIMESTAMP) FROM SYSIBM.SYSDUMMY1`, nil
}

// ListLockInformationJsonStringSql DB2 列出所有锁
// DB2 不支持 LIMIT，使用 FETCH FIRST n ROWS ONLY
func (x *DB2SqlProvider) ListLockInformationJsonStringSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT json_data FROM ` + tableFullName + ` FETCH FIRST 10000 ROWS ONLY`, nil
}