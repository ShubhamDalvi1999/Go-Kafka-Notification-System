package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"kafka-notify/pkg/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	GetNotificationByID(ctx context.Context, notificationID uuid.UUID) (*models.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAsDelivered(ctx context.Context, notificationID uuid.UUID) error
	MarkAsSent(ctx context.Context, notificationID uuid.UUID) error
	GetUnpublishedOutbox(ctx context.Context, limit int) ([]models.OutboxNotification, error)
	MarkOutboxPublished(ctx context.Context, outboxID int64) error
	CreateOutboxEntry(ctx context.Context, outboxItem *models.OutboxNotification) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserNotificationPreferences, error)
	UpdateUserPreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserNotificationPreferences) error
	GetUserEngagementStreak(ctx context.Context, userID uuid.UUID, streakType string) (*models.UserEngagementStreak, error)
	UpdateUserEngagementStreak(ctx context.Context, streak *models.UserEngagementStreak) error
	GetNotificationsByStatus(ctx context.Context, status models.DeliveryStatus, limit int) ([]models.Notification, error)
	GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]models.Notification, error)
	CreateDeliveryAttempt(ctx context.Context, attempt *models.NotificationDeliveryAttempt) error
	GetNotificationTemplates(ctx context.Context, notificationType models.NotificationType, channel models.NotificationChannel) ([]models.NotificationTemplate, error)
}

// PostgresNotificationRepository implements NotificationRepository using PostgreSQL
type PostgresNotificationRepository struct {
	db *sql.DB
}

// NewPostgresNotificationRepository creates a new PostgreSQL notification repository
func NewPostgresNotificationRepository(db *sql.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

// CreateNotification creates a new notification in the database
func (r *PostgresNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, type, channel, priority, template_id, title, message, 
			metadata, dedupe_key, scheduled_for, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Channel,
		notification.Priority,
		notification.TemplateID,
		notification.Title,
		notification.Message,
		notification.Metadata, // JSONMap handles JSON serialization automatically
		notification.DedupeKey,
		notification.ScheduledFor,
		notification.Status,
		notification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// GetUserNotifications retrieves notifications for a specific user
func (r *PostgresNotificationRepository) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, priority, template_id, title, message,
			   metadata, dedupe_key, created_at, scheduled_for, sent_at, delivered_at, read_at, status
		FROM notifications 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query user notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Priority, &n.TemplateID,
			&n.Title, &n.Message, &n.Metadata, &n.DedupeKey, &n.CreatedAt,
			&n.ScheduledFor, &n.SentAt, &n.DeliveredAt, &n.ReadAt, &n.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// GetNotificationByID retrieves a notification by its ID
func (r *PostgresNotificationRepository) GetNotificationByID(ctx context.Context, notificationID uuid.UUID) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, priority, template_id, title, message,
			   metadata, dedupe_key, created_at, scheduled_for, sent_at, delivered_at, read_at, status
		FROM notifications 
		WHERE id = $1
	`

	var n models.Notification
	err := r.db.QueryRowContext(ctx, query, notificationID).Scan(
		&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Priority, &n.TemplateID,
		&n.Title, &n.Message, &n.Metadata, &n.DedupeKey, &n.CreatedAt,
		&n.ScheduledFor, &n.SentAt, &n.DeliveredAt, &n.ReadAt, &n.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found: %s", notificationID)
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &n, nil
}

// MarkAsRead marks a notification as read
func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	query := `
		UPDATE notifications 
		SET read_at = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, models.StatusRead, now, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// MarkAsDelivered marks a notification as delivered
func (r *PostgresNotificationRepository) MarkAsDelivered(ctx context.Context, notificationID uuid.UUID) error {
	query := `
		UPDATE notifications 
		SET delivered_at = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, models.StatusDelivered, now, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as delivered: %w", err)
	}

	return nil
}

// MarkAsSent marks a notification as sent
func (r *PostgresNotificationRepository) MarkAsSent(ctx context.Context, notificationID uuid.UUID) error {
	query := `
		UPDATE notifications 
		SET sent_at = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, models.StatusSent, now, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return nil
}

// GetUnpublishedOutbox retrieves unpublished notifications from the outbox
func (r *PostgresNotificationRepository) GetUnpublishedOutbox(ctx context.Context, limit int) ([]models.OutboxNotification, error) {
	query := `
		SELECT id, notification_id, topic, payload, published, created_at, published_at
		FROM outbox_notifications 
		WHERE published = false 
		ORDER BY created_at ASC 
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpublished outbox: %w", err)
	}
	defer rows.Close()

	var outboxItems []models.OutboxNotification
	for rows.Next() {
		var item models.OutboxNotification
		err := rows.Scan(
			&item.ID, &item.NotificationID, &item.Topic, &item.Payload,
			&item.Published, &item.CreatedAt, &item.PublishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox item: %w", err)
		}
		outboxItems = append(outboxItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating outbox items: %w", err)
	}

	return outboxItems, nil
}

// MarkOutboxPublished marks an outbox item as published
func (r *PostgresNotificationRepository) MarkOutboxPublished(ctx context.Context, outboxID int64) error {
	query := `
		UPDATE outbox_notifications 
		SET published = true, published_at = $1
		WHERE id = $2
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, now, outboxID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox as published: %w", err)
	}

	return nil
}

// CreateOutboxEntry creates a new outbox entry
func (r *PostgresNotificationRepository) CreateOutboxEntry(ctx context.Context, outboxItem *models.OutboxNotification) error {
	query := `
		INSERT INTO outbox_notifications (
			notification_id, topic, payload, published, created_at
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		outboxItem.NotificationID,
		outboxItem.Topic,
		outboxItem.Payload, // JSONMap handles JSON serialization automatically
		outboxItem.Published,
		outboxItem.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create outbox entry: %w", err)
	}

	return nil
}

// GetUserPreferences retrieves notification preferences for a user
func (r *PostgresNotificationRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserNotificationPreferences, error) {
	query := `
		SELECT id, user_id, type, channel, enabled, quiet_hours_start, quiet_hours_end,
			   max_per_day, last_sent_at, metadata, created_at, updated_at
		FROM user_notification_preferences 
		WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user preferences: %w", err)
	}
	defer rows.Close()

	var preferences []models.UserNotificationPreferences
	for rows.Next() {
		var pref models.UserNotificationPreferences
		err := rows.Scan(
			&pref.ID, &pref.UserID, &pref.Type, &pref.Channel, &pref.Enabled,
			&pref.QuietHoursStart, &pref.QuietHoursEnd, &pref.MaxPerDay,
			&pref.LastSentAt, &pref.Metadata, &pref.CreatedAt, &pref.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan preference: %w", err)
		}
		preferences = append(preferences, pref)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating preferences: %w", err)
	}

	return preferences, nil
}

// UpdateUserPreferences updates notification preferences for a user
func (r *PostgresNotificationRepository) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserNotificationPreferences) error {
	query := `
		INSERT INTO user_notification_preferences (
			user_id, type, channel, enabled, quiet_hours_start, quiet_hours_end,
			max_per_day, metadata, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (userID, type, channel) 
		DO UPDATE SET 
			enabled = EXCLUDED.enabled,
			quiet_hours_start = EXCLUDED.quiet_hours_start,
			quiet_hours_end = EXCLUDED.quiet_hours_end,
			max_per_day = EXCLUDED.max_per_day,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		userID, prefs.Type, prefs.Channel, prefs.Enabled,
		prefs.QuietHoursStart, prefs.QuietHoursEnd, prefs.MaxPerDay,
		prefs.Metadata, now, // JSONMap handles JSON serialization automatically
	)

	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// GetUserEngagementStreak retrieves engagement streak for a user
func (r *PostgresNotificationRepository) GetUserEngagementStreak(ctx context.Context, userID uuid.UUID, streakType string) (*models.UserEngagementStreak, error) {
	query := `
		SELECT id, user_id, streak_type, current_streak, longest_streak,
			   last_activity_date, streak_start_date, total_activities, timezone,
			   created_at, updated_at
		FROM user_engagement_streaks 
		WHERE user_id = $1 AND streak_type = $2
	`

	var streak models.UserEngagementStreak
	err := r.db.QueryRowContext(ctx, query, userID, streakType).Scan(
		&streak.ID, &streak.UserID, &streak.StreakType, &streak.CurrentStreak,
		&streak.LongestStreak, &streak.LastActivityDate, &streak.StreakStartDate,
		&streak.TotalActivities, &streak.Timezone, &streak.CreatedAt, &streak.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("streak not found for user %s and type %s", userID, streakType)
		}
		return nil, fmt.Errorf("failed to get user engagement streak: %w", err)
	}

	return &streak, nil
}

// UpdateUserEngagementStreak updates or creates an engagement streak
func (r *PostgresNotificationRepository) UpdateUserEngagementStreak(ctx context.Context, streak *models.UserEngagementStreak) error {
	query := `
		INSERT INTO user_engagement_streaks (
			user_id, streak_type, current_streak, longest_streak,
			last_activity_date, streak_start_date, total_activities, timezone, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, streak_type) 
		DO UPDATE SET 
			current_streak = EXCLUDED.current_streak,
			longest_streak = EXCLUDED.longest_streak,
			last_activity_date = EXCLUDED.last_activity_date,
			streak_start_date = EXCLUDED.streak_start_date,
			total_activities = EXCLUDED.total_activities,
			timezone = EXCLUDED.timezone,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		streak.UserID, streak.StreakType, streak.CurrentStreak, streak.LongestStreak,
		streak.LastActivityDate, streak.StreakStartDate, streak.TotalActivities,
		streak.Timezone, now,
	)

	if err != nil {
		return fmt.Errorf("failed to update user engagement streak: %w", err)
	}

	return nil
}

// GetNotificationsByStatus retrieves notifications by their delivery status
func (r *PostgresNotificationRepository) GetNotificationsByStatus(ctx context.Context, status models.DeliveryStatus, limit int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, priority, template_id, title, message,
			   metadata, dedupe_key, created_at, scheduled_for, sent_at, delivered_at, read_at, status
		FROM notifications 
		WHERE status = $1 
		ORDER BY created_at ASC 
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications by status: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Priority, &n.TemplateID,
			&n.Title, &n.Message, &n.Metadata, &n.DedupeKey, &n.CreatedAt,
			&n.ScheduledFor, &n.SentAt, &n.DeliveredAt, &n.ReadAt, &n.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// GetScheduledNotifications retrieves notifications scheduled to be sent before a specific time
func (r *PostgresNotificationRepository) GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, priority, template_id, title, message,
			   metadata, dedupe_key, created_at, scheduled_for, sent_at, delivered_at, read_at, status
		FROM notifications 
		WHERE scheduled_for IS NOT NULL 
		  AND scheduled_for <= $1 
		  AND status = $2
		ORDER BY scheduled_for ASC 
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, before, models.StatusQueued, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query scheduled notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Priority, &n.TemplateID,
			&n.Title, &n.Message, &n.Metadata, &n.DedupeKey, &n.CreatedAt,
			&n.ScheduledFor, &n.SentAt, &n.DeliveredAt, &n.ReadAt, &n.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scheduled notifications: %w", err)
	}

	return notifications, nil
}

// CreateDeliveryAttempt creates a new delivery attempt record
func (r *PostgresNotificationRepository) CreateDeliveryAttempt(ctx context.Context, attempt *models.NotificationDeliveryAttempt) error {
	query := `
		INSERT INTO notification_delivery_attempts (
			notification_id, attempt_no, status, error_code, error_message,
			provider_message_id, latency_ms, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		attempt.NotificationID, attempt.AttemptNo, attempt.Status,
		attempt.ErrorCode, attempt.ErrorMessage, attempt.ProviderMessageID,
		attempt.LatencyMs, attempt.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create delivery attempt: %w", err)
	}

	return nil
}

// GetNotificationTemplates retrieves notification templates by type and channel
func (r *PostgresNotificationRepository) GetNotificationTemplates(ctx context.Context, notificationType models.NotificationType, channel models.NotificationChannel) ([]models.NotificationTemplate, error) {
	query := `
		SELECT id, type, channel, title, body, locale, priority, is_active, version, created_at
		FROM notification_templates 
		WHERE type = $1 AND channel = $2 AND is_active = true
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, notificationType, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification templates: %w", err)
	}
	defer rows.Close()

	var templates []models.NotificationTemplate
	for rows.Next() {
		var t models.NotificationTemplate
		err := rows.Scan(
			&t.ID, &t.Type, &t.Channel, &t.Title, &t.Body, &t.Locale,
			&t.Priority, &t.IsActive, &t.Version, &t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}
