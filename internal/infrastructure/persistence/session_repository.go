// Package persistence provides repository implementations
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common repository errors
var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
)

// SessionRepository handles session persistence
type SessionRepository struct {
	db *Database
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *Database) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session
func (r *SessionRepository) Create(ctx context.Context, session *SessionModel) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now().UTC()
	session.UpdatedAt = session.CreatedAt

	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a session by ID
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*SessionModel, error) {
	var session SessionModel
	err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// Update updates a session
func (r *SessionRepository) Update(ctx context.Context, session *SessionModel) error {
	session.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Save(session)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// UpdateState updates only the session state
func (r *SessionRepository) UpdateState(ctx context.Context, id, state string) error {
	result := r.db.WithContext(ctx).Model(&SessionModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// Close marks a session as closed
func (r *SessionRepository) Close(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&SessionModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"state":      "closed",
			"closed_at":  now,
			"updated_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// Delete soft-deletes a session
func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&SessionModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// List lists sessions with pagination
func (r *SessionRepository) List(ctx context.Context, opts *ListOptions) ([]SessionModel, int64, error) {
	var sessions []SessionModel
	var total int64

	query := r.db.WithContext(ctx).Model(&SessionModel{})

	// Apply filters
	if opts != nil {
		if opts.State != "" {
			query = query.Where("state = ?", opts.State)
		}
		if opts.ClientName != "" {
			query = query.Where("client_name ILIKE ?", "%"+opts.ClientName+"%")
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

	if err := query.Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// ListActive lists all active sessions
func (r *SessionRepository) ListActive(ctx context.Context) ([]SessionModel, error) {
	var sessions []SessionModel
	err := r.db.WithContext(ctx).
		Where("state IN ?", []string{"created", "initializing", "ready"}).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// CountByState counts sessions by state
func (r *SessionRepository) CountByState(ctx context.Context) (map[string]int64, error) {
	type result struct {
		State string
		Count int64
	}

	var results []result
	err := r.db.WithContext(ctx).
		Model(&SessionModel{}).
		Select("state, COUNT(*) as count").
		Group("state").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.State] = r.Count
	}
	return counts, nil
}

// CleanupOldSessions deletes sessions older than the specified duration
func (r *SessionRepository) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan)
	result := r.db.WithContext(ctx).
		Where("state = ? AND closed_at < ?", "closed", cutoff).
		Delete(&SessionModel{})
	return result.RowsAffected, result.Error
}

// ListOptions holds list query options
type ListOptions struct {
	Limit      int
	Offset     int
	OrderBy    string
	State      string
	ClientName string
	Since      time.Time
	Until      time.Time
}
