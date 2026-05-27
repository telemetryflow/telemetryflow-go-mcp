package services

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"gorm.io/gorm"
)

type DefaultDBProvider struct {
	gormDB *gorm.DB
	chConn driver.Conn
	chDB   string
	hasPG  bool
	hasCH  bool
}

func NewDefaultDBProvider(gormDB *gorm.DB, chConn driver.Conn, chDB string) *DefaultDBProvider {
	return &DefaultDBProvider{
		gormDB: gormDB,
		chConn: chConn,
		chDB:   chDB,
		hasPG:  gormDB != nil,
		hasCH:  chConn != nil,
	}
}

func (p *DefaultDBProvider) GormDB() *gorm.DB {
	return p.gormDB
}

func (p *DefaultDBProvider) ClickHouseConn() driver.Conn {
	return p.chConn
}

func (p *DefaultDBProvider) ClickHouseDB() string {
	return p.chDB
}

func (p *DefaultDBProvider) HasClickHouse() bool {
	return p.hasCH
}

func (p *DefaultDBProvider) HasPostgres() bool {
	return p.hasPG
}
