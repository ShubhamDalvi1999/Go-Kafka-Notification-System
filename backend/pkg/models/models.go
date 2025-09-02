package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============== ENUM TYPES ==============

// JSONMap is a custom type that can handle JSONB database fields
type JSONMap map[string]interface{}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type NotificationType string
type NotificationChannel string
type DeliveryStatus string
type PriorityLevel string

const (
	// Notification Types
	DailyReminder     NotificationType = "daily_reminder"
	StreakReminder    NotificationType = "streak_reminder"
	LastChanceAlert   NotificationType = "last_chance_alert"
	AchievementUnlock NotificationType = "achievement_unlock"
	XPGoalReminder    NotificationType = "xp_goal_reminder"
	LeagueUpdate      NotificationType = "league_update"
	WeMissYou         NotificationType = "we_miss_you"
	EventNotification NotificationType = "event_notification"
	NewCourse         NotificationType = "new_course"
	PracticeNeeded    NotificationType = "practice_needed"
	WeeklyRecap       NotificationType = "weekly_recap"

	// Notification Channels
	ChannelInApp NotificationChannel = "in_app"
	ChannelPush  NotificationChannel = "push"
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"

	// Delivery Status
	StatusQueued     DeliveryStatus = "queued"
	StatusSent       DeliveryStatus = "sent"
	StatusDelivered  DeliveryStatus = "delivered"
	StatusFailed     DeliveryStatus = "failed"
	StatusSuppressed DeliveryStatus = "suppressed"
	StatusRead       DeliveryStatus = "read"

	// Priority Levels
	PriorityLow    PriorityLevel = "low"
	PriorityMedium PriorityLevel = "medium"
	PriorityHigh   PriorityLevel = "high"
	PriorityUrgent PriorityLevel = "urgent"
)

// ============== CORE MODELS ==============

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	TotalXP   int       `json:"total_xp" db:"total_xp"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile represents extended user profile information
type UserProfile struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"id"`
	FullName  string    `json:"full_name" db:"full_name"`
	AvatarURL *string   `json:"avatar_url" db:"avatar_url"`
	Bio       *string   `json:"bio" db:"bio"`
	Username  *string   `json:"username" db:"username"`
	Location  *string   `json:"location" db:"location"`
	Website   *string   `json:"website" db:"website"`
	Skills    []string  `json:"skills" db:"skills"`
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Notification represents a notification record
type Notification struct {
	ID           uuid.UUID           `json:"id" db:"id"`
	UserID       uuid.UUID           `json:"user_id" db:"user_id"`
	Type         NotificationType    `json:"type" db:"type"`
	Channel      NotificationChannel `json:"channel" db:"channel"`
	Priority     PriorityLevel       `json:"priority" db:"priority"`
	TemplateID   *int64              `json:"template_id" db:"template_id"`
	Title        *string             `json:"title" db:"title"`
	Message      string              `json:"message" db:"message"`
	Metadata     JSONMap             `json:"metadata" db:"metadata"`
	DedupeKey    *string             `json:"dedupe_key" db:"dedupe_key"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	ScheduledFor *time.Time          `json:"scheduled_for" db:"scheduled_for"`
	SentAt       *time.Time          `json:"sent_at" db:"sent_at"`
	DeliveredAt  *time.Time          `json:"delivered_at" db:"delivered_at"`
	ReadAt       *time.Time          `json:"read_at" db:"read_at"`
	Status       DeliveryStatus      `json:"status" db:"status"`
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID        int64               `json:"id" db:"id"`
	Type      NotificationType    `json:"type" db:"type"`
	Channel   NotificationChannel `json:"channel" db:"channel"`
	Title     *string             `json:"title" db:"title"`
	Body      string              `json:"body" db:"body"`
	Locale    string              `json:"locale" db:"locale"`
	Priority  PriorityLevel       `json:"priority" db:"priority"`
	IsActive  bool                `json:"is_active" db:"is_active"`
	Version   int                 `json:"version" db:"version"`
	CreatedAt time.Time           `json:"created_at" db:"created_at"`
}

// UserNotificationPreferences represents user notification preferences
type UserNotificationPreferences struct {
	ID              int64               `json:"id" db:"id"`
	UserID          uuid.UUID           `json:"user_id" db:"user_id"`
	Type            NotificationType    `json:"type" db:"type"`
	Channel         NotificationChannel `json:"channel" db:"channel"`
	Enabled         bool                `json:"enabled" db:"enabled"`
	QuietHoursStart *string             `json:"quiet_hours_start" db:"quiet_hours_start"`
	QuietHoursEnd   *string             `json:"quiet_hours_end" db:"quiet_hours_end"`
	MaxPerDay       *int                `json:"max_per_day" db:"max_per_day"`
	LastSentAt      *time.Time          `json:"last_sent_at" db:"last_sent_at"`
	Metadata        JSONMap             `json:"metadata" db:"metadata"`
	CreatedAt       time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at" db:"updated_at"`
}

// NotificationDeliveryAttempt represents a delivery attempt
type NotificationDeliveryAttempt struct {
	ID                int64          `json:"id" db:"id"`
	NotificationID    uuid.UUID      `json:"notification_id" db:"notification_id"`
	AttemptNo         int            `json:"attempt_no" db:"attempt_no"`
	Status            DeliveryStatus `json:"status" db:"status"`
	ErrorCode         *string        `json:"error_code" db:"error_code"`
	ErrorMessage      *string        `json:"error_message" db:"error_message"`
	ProviderMessageID *string        `json:"provider_message_id" db:"provider_message_id"`
	LatencyMs         *int           `json:"latency_ms" db:"latency_ms"`
	CreatedAt         time.Time      `json:"created_at" db:"created_at"`
}

// OutboxNotification represents a notification in the outbox for Kafka
type OutboxNotification struct {
	ID             int64      `json:"id" db:"id"`
	NotificationID uuid.UUID  `json:"notification_id" db:"notification_id"`
	Topic          string     `json:"topic" db:"topic"`
	Payload        JSONMap    `json:"payload" db:"payload"`
	Published      bool       `json:"published" db:"published"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	PublishedAt    *time.Time `json:"published_at" db:"published_at"`
}

// UserEngagementStreak represents user engagement streaks
type UserEngagementStreak struct {
	ID               int64      `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	StreakType       string     `json:"streak_type" db:"streak_type"`
	CurrentStreak    int        `json:"current_streak" db:"current_streak"`
	LongestStreak    int        `json:"longest_streak" db:"longest_streak"`
	LastActivityDate *time.Time `json:"last_activity_date" db:"last_activity_date"`
	StreakStartDate  *time.Time `json:"streak_start_date" db:"streak_start_date"`
	TotalActivities  int        `json:"total_activities" db:"total_activities"`
	Timezone         string     `json:"timezone" db:"timezone"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// ============== REQUEST/RESPONSE MODELS ==============

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID       uuid.UUID           `json:"user_id" binding:"required"`
	Type         NotificationType    `json:"type" binding:"required"`
	Channel      NotificationChannel `json:"channel" binding:"required"`
	Priority     PriorityLevel       `json:"priority"`
	Title        *string             `json:"title"`
	Message      string              `json:"message" binding:"required"`
	Metadata     JSONMap             `json:"metadata"`
	ScheduledFor *time.Time          `json:"scheduled_for"`
}

// UpdateNotificationRequest represents a request to update a notification
type UpdateNotificationRequest struct {
	Status      *DeliveryStatus `json:"status"`
	SentAt      *time.Time      `json:"sent_at"`
	DeliveredAt *time.Time      `json:"delivered_at"`
	ReadAt      *time.Time      `json:"read_at"`
	Metadata    JSONMap         `json:"metadata"`
}

// NotificationPreferencesRequest represents a request to update notification preferences
type NotificationPreferencesRequest struct {
	Type            NotificationType    `json:"type" binding:"required"`
	Channel         NotificationChannel `json:"channel" binding:"required"`
	Enabled         bool                `json:"enabled"`
	QuietHoursStart *string             `json:"quiet_hours_start"`
	QuietHoursEnd   *string             `json:"quiet_hours_end"`
	MaxPerDay       *int                `json:"max_per_day"`
}

// ============== HELPER METHODS ==============

// IsRead returns true if the notification has been read
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}

// IsDelivered returns true if the notification has been delivered
func (n *Notification) IsDelivered() bool {
	return n.DeliveredAt != nil
}

// IsSent returns true if the notification has been sent
func (n *Notification) IsSent() bool {
	return n.SentAt != nil
}

// GetPriority returns the priority level as an integer for sorting
func (p PriorityLevel) GetPriority() int {
	switch p {
	case PriorityLow:
		return 1
	case PriorityMedium:
		return 2
	case PriorityHigh:
		return 3
	case PriorityUrgent:
		return 4
	default:
		return 0
	}
}

// IsValidNotificationType checks if the notification type is valid
func IsValidNotificationType(nt NotificationType) bool {
	validTypes := []NotificationType{
		DailyReminder, StreakReminder, LastChanceAlert, AchievementUnlock,
		XPGoalReminder, LeagueUpdate, WeMissYou, EventNotification,
		NewCourse, PracticeNeeded, WeeklyRecap,
	}

	for _, validType := range validTypes {
		if nt == validType {
			return true
		}
	}
	return false
}

// IsValidChannel checks if the notification channel is valid
func IsValidChannel(nc NotificationChannel) bool {
	validChannels := []NotificationChannel{
		ChannelInApp, ChannelPush, ChannelEmail, ChannelSMS,
	}

	for _, validChannel := range validChannels {
		if nc == validChannel {
			return true
		}
	}
	return false
}
