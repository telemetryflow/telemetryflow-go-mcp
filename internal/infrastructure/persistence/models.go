// Package persistence provides database models for GORM
package persistence

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// SessionModel represents a session in the database
type SessionModel struct {
	ID              string         `gorm:"type:uuid;primaryKey"`
	ProtocolVersion string         `gorm:"type:varchar(50);not null;default:'2024-11-05'"`
	State           string         `gorm:"type:varchar(50);not null;index"`
	ClientName      string         `gorm:"type:varchar(255)"`
	ClientVersion   string         `gorm:"type:varchar(50)"`
	ServerName      string         `gorm:"type:varchar(255);not null;default:'TelemetryFlow-MCP'"`
	ServerVersion   string         `gorm:"type:varchar(50);not null;default:'1.1.2'"`
	Capabilities    JSONB          `gorm:"type:jsonb"`
	LogLevel        string         `gorm:"type:varchar(50);default:'info'"`
	Metadata        JSONB          `gorm:"type:jsonb"`
	CreatedAt       time.Time      `gorm:"not null;index"`
	UpdatedAt       time.Time      `gorm:"not null"`
	ClosedAt        *time.Time     `gorm:"index"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for SessionModel
func (SessionModel) TableName() string {
	return "sessions"
}

// ConversationModel represents a conversation in the database
type ConversationModel struct {
	ID            string         `gorm:"type:uuid;primaryKey"`
	SessionID     string         `gorm:"type:uuid;not null;index"`
	Model         string         `gorm:"type:varchar(100);not null;index"`
	SystemPrompt  string         `gorm:"type:text"`
	Status        string         `gorm:"type:varchar(50);not null;index"`
	MaxTokens     int            `gorm:"not null;default:4096"`
	Temperature   float64        `gorm:"not null;default:1.0"`
	TopP          float64        `gorm:"not null;default:1.0"`
	TopK          int            `gorm:"default:0"`
	StopSequences JSONB          `gorm:"type:jsonb"`
	Metadata      JSONB          `gorm:"type:jsonb"`
	CreatedAt     time.Time      `gorm:"not null;index"`
	UpdatedAt     time.Time      `gorm:"not null"`
	ClosedAt      *time.Time     `gorm:"index"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relations
	Session  *SessionModel  `gorm:"foreignKey:SessionID;references:ID"`
	Messages []MessageModel `gorm:"foreignKey:ConversationID;references:ID"`
}

// TableName returns the table name for ConversationModel
func (ConversationModel) TableName() string {
	return "conversations"
}

// MessageModel represents a message in the database
type MessageModel struct {
	ID             string    `gorm:"type:uuid;primaryKey"`
	ConversationID string    `gorm:"type:uuid;not null;index"`
	Role           string    `gorm:"type:varchar(50);not null;index"`
	Content        JSONB     `gorm:"type:jsonb;not null"`
	TokenCount     int       `gorm:"default:0"`
	CreatedAt      time.Time `gorm:"not null;index"`

	// Relations
	Conversation *ConversationModel `gorm:"foreignKey:ConversationID;references:ID"`
}

// TableName returns the table name for MessageModel
func (MessageModel) TableName() string {
	return "messages"
}

// ToolModel represents a tool definition in the database
type ToolModel struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string         `gorm:"type:text"`
	InputSchema JSONB          `gorm:"type:jsonb"`
	Category    string         `gorm:"type:varchar(100);index"`
	Tags        JSONB          `gorm:"type:jsonb"`
	IsEnabled   bool           `gorm:"not null;default:true;index"`
	RateLimit   JSONB          `gorm:"type:jsonb"`
	Timeout     int            `gorm:"default:30"` // in seconds
	Metadata    JSONB          `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for ToolModel
func (ToolModel) TableName() string {
	return "tools"
}

// ResourceModel represents a resource definition in the database
type ResourceModel struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	URI         string         `gorm:"type:varchar(2048);uniqueIndex;not null"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	MimeType    string         `gorm:"type:varchar(255)"`
	IsTemplate  bool           `gorm:"not null;default:false"`
	URITemplate string         `gorm:"type:varchar(2048)"`
	Annotations JSONB          `gorm:"type:jsonb"`
	Metadata    JSONB          `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for ResourceModel
func (ResourceModel) TableName() string {
	return "resources"
}

// PromptModel represents a prompt definition in the database
type PromptModel struct {
	ID          string         `gorm:"type:uuid;primaryKey"`
	Name        string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string         `gorm:"type:text"`
	Arguments   JSONB          `gorm:"type:jsonb"`
	Metadata    JSONB          `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for PromptModel
func (PromptModel) TableName() string {
	return "prompts"
}

// ToolCallModel represents a tool call record in the database
type ToolCallModel struct {
	ID             string     `gorm:"type:uuid;primaryKey"`
	SessionID      string     `gorm:"type:uuid;index"`
	ConversationID string     `gorm:"type:uuid;index"`
	MessageID      string     `gorm:"type:uuid;index"`
	ToolName       string     `gorm:"type:varchar(255);not null;index"`
	Input          JSONB      `gorm:"type:jsonb"`
	Output         JSONB      `gorm:"type:jsonb"`
	IsError        bool       `gorm:"not null;default:false;index"`
	DurationMs     int64      `gorm:"not null"`
	StartedAt      time.Time  `gorm:"not null;index"`
	CompletedAt    *time.Time `gorm:"index"`
}

// TableName returns the table name for ToolCallModel
func (ToolCallModel) TableName() string {
	return "tool_calls"
}

// APIRequestModel represents an API request record in the database
type APIRequestModel struct {
	ID             string     `gorm:"type:uuid;primaryKey"`
	SessionID      string     `gorm:"type:uuid;index"`
	ConversationID string     `gorm:"type:uuid;index"`
	Model          string     `gorm:"type:varchar(100);index"`
	InputTokens    int        `gorm:"not null;default:0"`
	OutputTokens   int        `gorm:"not null;default:0"`
	TotalTokens    int        `gorm:"not null;default:0"`
	DurationMs     int64      `gorm:"not null"`
	StatusCode     int        `gorm:"index"`
	ErrorMessage   string     `gorm:"type:text"`
	StartedAt      time.Time  `gorm:"not null;index"`
	CompletedAt    *time.Time `gorm:"index"`
}

// TableName returns the table name for APIRequestModel
func (APIRequestModel) TableName() string {
	return "api_requests"
}

// AuditLogModel represents an audit log entry in the database
type AuditLogModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	SessionID string    `gorm:"type:uuid;index"`
	Action    string    `gorm:"type:varchar(100);not null;index"`
	Resource  string    `gorm:"type:varchar(255);index"`
	Details   JSONB     `gorm:"type:jsonb"`
	UserAgent string    `gorm:"type:varchar(500)"`
	IPAddress string    `gorm:"type:varchar(45)"`
	CreatedAt time.Time `gorm:"not null;index"`
}

// TableName returns the table name for AuditLogModel
func (AuditLogModel) TableName() string {
	return "audit_logs"
}

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value returns the JSON value for database storage
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan scans a JSON value from the database
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSONB value")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// JSONBArray is a custom type for PostgreSQL JSONB array columns
type JSONBArray []interface{}

// Value returns the JSON value for database storage
func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan scans a JSON value from the database
func (j *JSONBArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSONBArray value")
	}

	var result []interface{}
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// AllModels returns all database models for migration
func AllModels() []interface{} {
	return []interface{}{
		&SessionModel{},
		&ConversationModel{},
		&MessageModel{},
		&ToolModel{},
		&ResourceModel{},
		&PromptModel{},
		&ToolCallModel{},
		&APIRequestModel{},
		&AuditLogModel{},
	}
}
