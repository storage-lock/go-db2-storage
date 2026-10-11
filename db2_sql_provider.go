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
//
// 注意：列名必须与 Sql92Provider 未被本 Provider 覆盖的
// CreateWithVersionSql/UpdateWithVersionSql/DeleteWithVersionSql/FindLockInformationJsonStringByIdSql
// 保持一致（lock_id/owner_id/version/lock_information_json_string）——这几个方法本 Provider
// 没有重写，会直接继承基类实现，如果表结构的列名对不上，CAS 相关的每条 SQL 都会在真实 DB2 上
// 报"列不存在"（SQLSTATE 42703）。修复前这里曾经是 lock_id/version/json_data/created_at，
// 缺 owner_id 且 JSON 列名对不上，导致本存储在真实 DB2 上完全不可用（未接入真实驱动测试，
// 靠走读代码发现）。
func (x *DB2SqlProvider) CreateTableSql(ctx context.Context, tableFullName string) (string, []any) {
	sql := `CREATE TABLE ` + tableFullName + ` (
		lock_id VARCHAR(255) PRIMARY KEY NOT NULL,
		owner_id VARCHAR(255) NOT NULL,
		version BIGINT NOT NULL,
		lock_information_json_string CLOB(16M) NOT NULL
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
// DB2 不支持 LIMIT，使用 FETCH FIRST n ROWS ONLY。
// 修复：原来查询的是不存在的 json_data 列（与 CreateTableSql 的列名不一致，见上），改为实际的
// lock_information_json_string 列。
func (x *DB2SqlProvider) ListLockInformationJsonStringSql(ctx context.Context, tableFullName string) (string, []any) {
	return `SELECT lock_information_json_string FROM ` + tableFullName + ` FETCH FIRST 10000 ROWS ONLY`, nil
}