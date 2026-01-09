// Package persistence provides repository implementations
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationRepository handles conversation persistence
type ConversationRepository struct {
	db *Database
}

// NewConversationRepository creates a new ConversationRepository
func NewConversationRepository(db *Database) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// Create creates a new conversation
func (r *ConversationRepository) Create(ctx context.Context, conversation *ConversationModel) error {
	if conversation.ID == "" {
		conversation.ID = uuid.New().String()
	}
	conversation.CreatedAt = time.Now().UTC()
	conversation.UpdatedAt = conversation.CreatedAt

	return r.db.WithContext(ctx).Create(conversation).Error
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepository) GetByID(ctx context.Context, id string) (*ConversationModel, error) {
	var conversation ConversationModel
	err := r.db.WithContext(ctx).First(&conversation, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}
	return &conversation, nil
}

// GetByIDWithMessages retrieves a conversation with all messages
func (r *ConversationRepository) GetByIDWithMessages(ctx context.Context, id string) (*ConversationModel, error) {
	var conversation ConversationModel
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&conversation, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConversationNotFound
		}
		return nil, err
	}
	return &conversation, nil
}

// Update updates a conversation
func (r *ConversationRepository) Update(ctx context.Context, conversation *ConversationModel) error {
	conversation.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Save(conversation)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// UpdateStatus updates only the conversation status
func (r *ConversationRepository) UpdateStatus(ctx context.Context, id, status string) error {
	result := r.db.WithContext(ctx).Model(&ConversationModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// Close marks a conversation as closed
func (r *ConversationRepository) Close(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&ConversationModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "closed",
			"closed_at":  now,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// Delete soft-deletes a conversation
func (r *ConversationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&ConversationModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrConversationNotFound
	}
	return nil
}

// ListBySession lists conversations for a session
func (r *ConversationRepository) ListBySession(ctx context.Context, sessionID string, opts *ListOptions) ([]ConversationModel, int64, error) {
	var conversations []ConversationModel
	var total int64

	query := r.db.WithContext(ctx).Model(&ConversationModel{}).Where("session_id = ?", sessionID)

	// Apply filters
	if opts != nil {
		if opts.State != "" {
			query = query.Where("status = ?", opts.State)
		}
		if !opts.Since.IsZero() {
			query = query.Where("created_at >= ?", opts.Since)
		}
		if !opts.Until.IsZero() {
			query = query.Where("created_at <= ?", opts.Until)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if opts != nil {
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}
		if opts.OrderBy != "" {
			query = query.Order(opts.OrderBy)
		} else {
			query = query.Order("created_at DESC")
		}
	}

	if err := query.Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

// ListActive lists all active conversations
func (r *ConversationRepository) ListActive(ctx context.Context) ([]ConversationModel, error) {
	var conversations []ConversationModel
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("created_at DESC").
		Find(&conversations).Error
	return conversations, err
}

// CountByModel counts conversations by model
func (r *ConversationRepository) CountByModel(ctx context.Context) (map[string]int64, error) {
	type result struct {
		Model string
		Count int64
	}

	var results []result
	err := r.db.WithContext(ctx).
		Model(&ConversationModel{}).
		Select("model, COUNT(*) as count").
		Group("model").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Model] = r.Count
	}
	return counts, nil
}

// GetMessageCount returns the number of messages in a conversation
func (r *ConversationRepository) GetMessageCount(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&MessageModel{}).
		Where("conversation_id = ?", conversationID).
		Count(&count).Error
	return count, err
}

// MessageRepository handles message persistence
type MessageRepository struct {
	db *Database
}

// NewMessageRepository creates a new MessageRepository
func NewMessageRepository(db *Database) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(ctx context.Context, message *MessageModel) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	message.CreatedAt = time.Now().UTC()

	return r.db.WithContext(ctx).Create(message).Error
}

// CreateBatch creates multiple messages
func (r *MessageRepository) CreateBatch(ctx context.Context, messages []MessageModel) error {
	now := time.Now().UTC()
	for i := range messages {
		if messages[i].ID == "" {
			messages[i].ID = uuid.New().String()
		}
		messages[i].CreatedAt = now
	}
	return r.db.WithContext(ctx).Create(&messages).Error
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(ctx context.Context, id string) (*MessageModel, error) {
	var message MessageModel
	err := r.db.WithContext(ctx).First(&message, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}
	return &message, nil
}

// ListByConversation lists messages for a conversation
func (r *MessageRepository) ListByConversation(ctx context.Context, conversationID string) ([]MessageModel, error) {
	var messages []MessageModel
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}

// GetLastMessages retrieves the last N messages for a conversation
func (r *MessageRepository) GetLastMessages(ctx context.Context, conversationID string, limit int) ([]MessageModel, error) {
	var messages []MessageModel
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountTokens returns the total token count for a conversation
func (r *MessageRepository) CountTokens(ctx context.Context, conversationID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&MessageModel{}).
		Where("conversation_id = ?", conversationID).
		Select("COALESCE(SUM(token_count), 0)").
		Scan(&total).Error
	return total, err
}

// DeleteByConversation deletes all messages for a conversation
func (r *MessageRepository) DeleteByConversation(ctx context.Context, conversationID string) error {
	return r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Delete(&MessageModel{}).Error
}
