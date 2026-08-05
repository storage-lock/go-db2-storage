package db2_storage

import (
	"context"
	"os"
	"testing"

	storage_test_helper "github.com/storage-lock/go-storage-test-helper"
	"github.com/stretchr/testify/assert"
)

func TestDB2Storage(t *testing.T) {
	dsn := os.Getenv("STORAGE_LOCK_DB2_DSN")
	if dsn == "" {
		t.Skip("STORAGE_LOCK_DB2_DSN not set, skipping DB2 test")
	}

	ctx := context.Background()
	cm := NewDB2ConnectionManagerFromDSN(dsn)
	s, err := NewDB2Storage(ctx, &DB2StorageOptions{
		ConnectionManager: cm,
		TableName:         storage_test_helper.TestTableName,
	})
	assert.Nil(t, err)
	storage_test_helper.TestStorage(t, s)
}
