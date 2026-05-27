package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/repositories"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

type GormSessionRepository struct {
	db *gorm.DB
}

func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

// Save stores a session in the database using GORM
// It takes a context and a session aggregate as input
// and returns an error if the operation fails
func (r *GormSessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
	// Convert the session aggregate to a GORM model
	model := sessionToModel(session)
	// Save the model to the database with the given context
	// and return any error that occurs
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GormSessionRepository) FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	var model SessionModel
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToSession(&model)
}

func (r *GormSessionRepository) FindAll(ctx context.Context) ([]*aggregates.Session, error) {
	var models []SessionModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	sessions := make([]*aggregates.Session, 0, len(models))
	for _, m := range models {
		s, err := modelToSession(&m)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *GormSessionRepository) FindActive(ctx context.Context) ([]*aggregates.Session, error) {
	var models []SessionModel
	if err := r.db.WithContext(ctx).Where("state IN ?", []string{"created", "initializing", "ready"}).Find(&models).Error; err != nil {
		return nil, err
	}
	sessions := make([]*aggregates.Session, 0, len(models))
	for _, m := range models {
		s, err := modelToSession(&m)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *GormSessionRepository) Delete(ctx context.Context, id vo.SessionID) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&SessionModel{}).Error
}

func (r *GormSessionRepository) Exists(ctx context.Context, id vo.SessionID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&SessionModel{}).Where("id = ?", id.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormSessionRepository) Count(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&SessionModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

var _ repositories.ISessionRepository = (*GormSessionRepository)(nil)

type GormConversationRepository struct {
	db *gorm.DB
}

func NewGormConversationRepository(db *gorm.DB) *GormConversationRepository {
	return &GormConversationRepository{db: db}
}

func (r *GormConversationRepository) Save(ctx context.Context, conv *aggregates.Conversation) error {
	model := conversationToModel(conv)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GormConversationRepository) FindByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	var model ConversationModel
	if err := r.db.WithContext(ctx).Preload("Messages").Where("id = ?", id.String()).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToConversation(&model)
}

func (r *GormConversationRepository) FindBySessionID(ctx context.Context, sessionID vo.SessionID) ([]*aggregates.Conversation, error) {
	var models []ConversationModel
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID.String()).Find(&models).Error; err != nil {
		return nil, err
	}
	conversations := make([]*aggregates.Conversation, 0, len(models))
	for _, m := range models {
		c, err := modelToConversation(&m)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, nil
}

func (r *GormConversationRepository) FindActive(ctx context.Context) ([]*aggregates.Conversation, error) {
	var models []ConversationModel
	if err := r.db.WithContext(ctx).Where("status = ?", "active").Find(&models).Error; err != nil {
		return nil, err
	}
	conversations := make([]*aggregates.Conversation, 0, len(models))
	for _, m := range models {
		c, err := modelToConversation(&m)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, nil
}

func (r *GormConversationRepository) Delete(ctx context.Context, id vo.ConversationID) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ConversationModel{}).Error
}

func (r *GormConversationRepository) Exists(ctx context.Context, id vo.ConversationID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("id = ?", id.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormConversationRepository) Count(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ConversationModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *GormConversationRepository) CountBySessionID(ctx context.Context, sessionID vo.SessionID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("session_id = ?", sessionID.String()).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

var _ repositories.IConversationRepository = (*GormConversationRepository)(nil)

type GormToolRepository struct {
	db *gorm.DB
}

func NewGormToolRepository(db *gorm.DB) *GormToolRepository {
	return &GormToolRepository{db: db}
}

func (r *GormToolRepository) Register(ctx context.Context, tool *entities.Tool) error {
	model := toolToModel(tool)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *GormToolRepository) Unregister(ctx context.Context, name vo.ToolName) error {
	return r.db.WithContext(ctx).Where("name = ?", name.String()).Delete(&ToolModel{}).Error
}

func (r *GormToolRepository) FindByName(ctx context.Context, name vo.ToolName) (*entities.Tool, error) {
	var model ToolModel
	if err := r.db.WithContext(ctx).Where("name = ?", name.String()).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return modelToTool(&model)
}

func (r *GormToolRepository) FindAll(ctx context.Context) ([]*entities.Tool, error) {
	var models []ToolModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	tools := make([]*entities.Tool, 0, len(models))
	for _, m := range models {
		t, err := modelToTool(&m)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *GormToolRepository) FindByCategory(ctx context.Context, category string) ([]*entities.Tool, error) {
	var models []ToolModel
	if err := r.db.WithContext(ctx).Where("category = ?", category).Find(&models).Error; err != nil {
		return nil, err
	}
	tools := make([]*entities.Tool, 0, len(models))
	for _, m := range models {
		t, err := modelToTool(&m)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *GormToolRepository) FindByTag(ctx context.Context, tag string) ([]*entities.Tool, error) {
	var models []ToolModel
	if err := r.db.WithContext(ctx).Where("tags::jsonb @> ?::jsonb", fmt.Sprintf(`["%s"]`, tag)).Find(&models).Error; err != nil {
		return nil, err
	}
	tools := make([]*entities.Tool, 0, len(models))
	for _, m := range models {
		t, err := modelToTool(&m)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *GormToolRepository) FindEnabled(ctx context.Context) ([]*entities.Tool, error) {
	var models []ToolModel
	if err := r.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&models).Error; err != nil {
		return nil, err
	}
	tools := make([]*entities.Tool, 0, len(models))
	for _, m := range models {
		t, err := modelToTool(&m)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

func (r *GormToolRepository) Exists(ctx context.Context, name vo.ToolName) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ToolModel{}).Where("name = ?", name.String()).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormToolRepository) Count(ctx context.Context) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&ToolModel{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

var _ repositories.IToolRepository = (*GormToolRepository)(nil)

func sessionToModel(s *aggregates.Session) *SessionModel {
	m := &SessionModel{
		ID:              s.ID().String(),
		ProtocolVersion: s.ProtocolVersion().String(),
		State:           string(s.State()),
		ServerName:      s.ServerInfo().Name,
		ServerVersion:   s.ServerInfo().Version,
		LogLevel:        s.LogLevel().String(),
		CreatedAt:       s.CreatedAt(),
		UpdatedAt:       s.UpdatedAt(),
		ClosedAt:        s.ClosedAt(),
	}
	if ci := s.ClientInfo(); ci != nil {
		m.ClientName = ci.Name
		m.ClientVersion = ci.Version
	}
	if caps := s.Capabilities(); caps != nil {
		b, _ := json.Marshal(caps)
		_ = json.Unmarshal(b, &m.Capabilities)
	}
	if len(s.Metadata()) > 0 {
		b, _ := json.Marshal(s.Metadata())
		_ = json.Unmarshal(b, &m.Metadata)
	}
	return m
}

func modelToSession(m *SessionModel) (*aggregates.Session, error) {
	sid, err := vo.NewSessionID(m.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID %q: %w", m.ID, err)
	}

	s := aggregates.RestoreSession(
		sid,
		vo.NewMCPProtocolVersion(m.ProtocolVersion),
		aggregates.SessionState(m.State),
		m.ServerName,
		m.ServerVersion,
		m.ClientName,
		m.ClientVersion,
		m.LogLevel,
		m.CreatedAt,
		m.UpdatedAt,
		m.ClosedAt,
	)
	return s, nil
}

func conversationToModel(c *aggregates.Conversation) *ConversationModel {
	m := &ConversationModel{
		ID:        c.ID().String(),
		SessionID: c.SessionID().String(),
		Model:     string(c.Model()),
		Status:    string(c.Status()),
		CreatedAt: c.CreatedAt(),
		UpdatedAt: c.UpdatedAt(),
		ClosedAt:  c.ClosedAt(),
	}
	return m
}

func modelToConversation(m *ConversationModel) (*aggregates.Conversation, error) {
	cid, err := vo.NewConversationID(m.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid conversation ID %q: %w", m.ID, err)
	}
	sid, err := vo.NewSessionID(m.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID %q: %w", m.SessionID, err)
	}

	c := aggregates.RestoreConversation(cid, sid, m.Model, m.Status, m.CreatedAt, m.UpdatedAt, m.ClosedAt)
	return c, nil
}

func toolToModel(t *entities.Tool) *ToolModel {
	m := &ToolModel{
		Name:        t.Name().String(),
		Description: t.Description().String(),
		Category:    t.Category(),
		IsEnabled:   t.IsEnabled(),
		Timeout:     int(t.Timeout().Seconds()),
		CreatedAt:   t.CreatedAt(),
		UpdatedAt:   t.UpdatedAt(),
	}
	if t.InputSchema() != nil {
		b, _ := json.Marshal(t.InputSchema())
		_ = json.Unmarshal(b, &m.InputSchema)
	}
	if len(t.Tags()) > 0 {
		b, _ := json.Marshal(t.Tags())
		_ = json.Unmarshal(b, &m.Tags)
	}
	return m
}

func modelToTool(m *ToolModel) (*entities.Tool, error) {
	name, err := vo.NewToolName(m.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid tool name %q: %w", m.Name, err)
	}
	desc, err := vo.NewToolDescription(m.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid tool description: %w", err)
	}

	var schema *entities.JSONSchema
	if m.InputSchema != nil {
		b, _ := json.Marshal(m.InputSchema)
		schema = &entities.JSONSchema{}
		_ = json.Unmarshal(b, schema)
	}

	t, err := entities.NewTool(name, desc, schema)
	if err != nil {
		return nil, err
	}
	t.SetCategory(m.Category)

	var tags []string
	if m.Tags != nil {
		b, _ := json.Marshal(m.Tags)
		_ = json.Unmarshal(b, &tags)
	}
	t.SetTags(tags)

	if !m.IsEnabled {
		t.Disable()
	}
	t.SetTimeout(time.Duration(m.Timeout) * time.Second)

	return t, nil
}
