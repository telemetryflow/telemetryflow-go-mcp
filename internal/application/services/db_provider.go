// Package services provides functionality for the TelemetryFlow GO MCP Server.
//
// TelemetryFlow GO MCP Server - Community Enterprise Observability Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
// Open Source Software built by Telemetri Data Indonesia.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
