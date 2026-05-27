package persistence_test

import (
	"context"
	"reflect"
	"testing"
	"unsafe"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
)

type mockClickHouseRows struct {
	nextCount  int
	maxRows    int
	scanValues [][]interface{}
	closeErr   error
	errVal     error
}

func (m *mockClickHouseRows) Next() bool {
	if m.nextCount < m.maxRows {
		m.nextCount++
		return true
	}
	return false
}

func (m *mockClickHouseRows) Scan(dest ...interface{}) error {
	rowIdx := m.nextCount - 1
	if rowIdx >= 0 && rowIdx < len(m.scanValues) {
		for i, val := range m.scanValues[rowIdx] {
			reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(val))
		}
	}
	return nil
}

func (m *mockClickHouseRows) ScanStruct(_ interface{}) error   { return nil }
func (m *mockClickHouseRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockClickHouseRows) Totals(_ ...interface{}) error    { return nil }
func (m *mockClickHouseRows) Columns() []string                { return nil }
func (m *mockClickHouseRows) Close() error                     { return m.closeErr }
func (m *mockClickHouseRows) Err() error                       { return m.errVal }

type mockClickHouseBatch struct {
	appendErr error
	sendErr   error
	abortErr  error
}

func (m *mockClickHouseBatch) Abort() error                     { return m.abortErr }
func (m *mockClickHouseBatch) Append(_ ...interface{}) error    { return m.appendErr }
func (m *mockClickHouseBatch) AppendStruct(_ interface{}) error { return nil }
func (m *mockClickHouseBatch) Column(_ int) driver.BatchColumn  { return nil }
func (m *mockClickHouseBatch) Flush() error                     { return nil }
func (m *mockClickHouseBatch) Send() error                      { return m.sendErr }
func (m *mockClickHouseBatch) IsSent() bool                     { return false }
func (m *mockClickHouseBatch) Rows() int                        { return 0 }
func (m *mockClickHouseBatch) Columns() []column.Interface      { return nil }
func (m *mockClickHouseBatch) Close() error                     { return nil }

type mockClickHouseConn struct {
	queryFn   func(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
	execFn    func(ctx context.Context, query string, args ...interface{}) error
	pingFn    func(ctx context.Context) error
	closeFn   func() error
	prepareFn func(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error)
}

func (m *mockClickHouseConn) Contributors() []string                        { return nil }
func (m *mockClickHouseConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockClickHouseConn) Select(_ context.Context, _ interface{}, _ string, _ ...interface{}) error {
	return nil
}
func (m *mockClickHouseConn) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	if m.queryFn != nil {
		return m.queryFn(ctx, query, args...)
	}
	return nil, nil
}
func (m *mockClickHouseConn) QueryRow(_ context.Context, _ string, _ ...interface{}) driver.Row {
	return nil
}
func (m *mockClickHouseConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	if m.prepareFn != nil {
		return m.prepareFn(ctx, query, opts...)
	}
	return &mockClickHouseBatch{}, nil
}
func (m *mockClickHouseConn) Exec(ctx context.Context, query string, args ...interface{}) error {
	if m.execFn != nil {
		return m.execFn(ctx, query, args...)
	}
	return nil
}
func (m *mockClickHouseConn) AsyncInsert(_ context.Context, _ string, _ bool, _ ...interface{}) error {
	return nil
}
func (m *mockClickHouseConn) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}
func (m *mockClickHouseConn) Stats() driver.Stats { return driver.Stats{} }
func (m *mockClickHouseConn) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

func setClickHouseConn(ch *persistence.ClickHouse, conn driver.Conn) {
	v := reflect.ValueOf(ch).Elem()
	f := v.FieldByName("conn")
	reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem().Set(reflect.ValueOf(conn))
}

func newClickHouseWithMockConn(t *testing.T, mock driver.Conn) *persistence.ClickHouse {
	t.Helper()
	ch := &persistence.ClickHouse{}
	setClickHouseConn(ch, mock)
	return ch
}
