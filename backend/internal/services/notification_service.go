package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"kafka-notify/pkg/models"
	"kafka-notify/pkg/repository"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// NotificationService defines the interface for notification operations
type NotificationService interface {
	CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error)
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, notificationID uuid.UUID) error
	UpdateUserPreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserNotificationPreferences) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserNotificationPreferences, error)
	CreateDailyReminder(ctx context.Context, user models.User) error
	CreateStreakReminder(ctx context.Context, user models.User) error
	ProcessOutbox(ctx context.Context) error
}

// notificationService implements NotificationService
type notificationService struct {
	repository repository.NotificationRepository
	producer   sarama.SyncProducer
	topic      string
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo repository.NotificationRepository, producer sarama.SyncProducer, topic string) NotificationService {
	return &notificationService{
		repository: repo,
		producer:   producer,
		topic:      topic,
	}
}

// CreateNotification creates a new notification
func (s *notificationService) CreateNotification(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	// Validate notification type
	if !models.IsValidNotificationType(req.Type) {
		return nil, fmt.Errorf("invalid notification type: %s", req.Type)
	}

	// Validate channel
	if !models.IsValidChannel(req.Channel) {
		return nil, fmt.Errorf("invalid notification channel: %s", req.Channel)
	}

	// Create notification
	notification := &models.Notification{
		ID:           uuid.New(),
		UserID:       req.UserID,
		Type:         req.Type,
		Channel:      req.Channel,
		Priority:     req.Priority,
		Title:        req.Title,
		Message:      req.Message,
		Metadata:     req.Metadata,
		Status:       models.StatusQueued,
		CreatedAt:    time.Now(),
		ScheduledFor: req.ScheduledFor,
	}

	// Save to database
	if err := s.repository.CreateNotification(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Create outbox entry for Kafka
	outboxItem := &models.OutboxNotification{
		NotificationID: notification.ID,
		Topic:          s.topic,
		Payload: models.JSONMap{
			"id":         notification.ID.String(),
			"user_id":    notification.UserID.String(),
			"type":       notification.Type,
			"channel":    notification.Channel,
			"priority":   notification.Priority,
			"title":      notification.Title,
			"message":    notification.Message,
			"created_at": notification.CreatedAt,
		},
		Published: false,
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreateOutboxEntry(ctx, outboxItem); err != nil {
		return nil, fmt.Errorf("failed to create outbox entry: %w", err)
	}

	// Immediate publish only if explicitly enabled (OUTBOX_IMMEDIATE_PUBLISH=true)
	if strings.EqualFold(os.Getenv("OUTBOX_IMMEDIATE_PUBLISH"), "true") {
		_ = s.ProcessOutbox(ctx)
	}

	return notification, nil
}

// GetUserNotifications retrieves notifications for a specific user
func (s *notificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repository.GetUserNotifications(ctx, userID, limit, offset)
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return s.repository.MarkAsRead(ctx, notificationID)
}

// UpdateUserPreferences updates notification preferences for a user
func (s *notificationService) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserNotificationPreferences) error {
	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()
	return s.repository.UpdateUserPreferences(ctx, userID, prefs)
}

// GetUserPreferences retrieves notification preferences for a user
func (s *notificationService) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserNotificationPreferences, error) {
	return s.repository.GetUserPreferences(ctx, userID)
}

// CreateDailyReminder creates a daily reminder for a user
func (s *notificationService) CreateDailyReminder(ctx context.Context, user models.User) error {
	// Get user engagement streak
	streak, err := s.repository.GetUserEngagementStreak(ctx, user.ID, "practice")
	if err != nil {
		// Continue with default streak value
	}

	currentStreak := 0
	if streak != nil {
		currentStreak = streak.CurrentStreak
	}

	// Create daily reminder notification
	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      models.DailyReminder,
		Channel:   models.ChannelInApp,
		Priority:  models.PriorityMedium,
		Title:     stringPtr("Time to Practice!"),
		Message:   fmt.Sprintf("Hey %s! It's time for your daily practice session. Keep your %d-day streak alive! 🔥", user.Name, currentStreak),
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repository.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create daily reminder: %w", err)
	}

	// Create outbox entry
	outboxItem := &models.OutboxNotification{
		NotificationID: notification.ID,
		Topic:          s.topic,
		Payload: map[string]interface{}{
			"id":         notification.ID.String(),
			"user_id":    notification.UserID.String(),
			"type":       notification.Type,
			"channel":    notification.Channel,
			"priority":   notification.Priority,
			"title":      notification.Title,
			"message":    notification.Message,
			"created_at": notification.CreatedAt,
		},
		Published: false,
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreateOutboxEntry(ctx, outboxItem); err != nil {
		return fmt.Errorf("failed to create outbox entry for daily reminder: %w", err)
	}

	return nil
}

// CreateStreakReminder creates a streak reminder for a user
func (s *notificationService) CreateStreakReminder(ctx context.Context, user models.User) error {
	// Get user engagement streak
	streak, err := s.repository.GetUserEngagementStreak(ctx, user.ID, "practice")
	if err != nil {
		return fmt.Errorf("failed to get user streak: %w", err)
	}

	if streak.CurrentStreak == 0 {
		return fmt.Errorf("user has no active streak")
	}

	// Create streak reminder notification
	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      models.StreakReminder,
		Channel:   models.ChannelInApp,
		Priority:  models.PriorityHigh,
		Title:     stringPtr("Don't Break Your Streak!"),
		Message:   fmt.Sprintf("%s, you haven't practiced today! Your %d-day streak is at risk. Practice now to keep it going!", user.Name, streak.CurrentStreak),
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repository.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create streak reminder: %w", err)
	}

	// Create outbox entry
	outboxItem := &models.OutboxNotification{
		NotificationID: notification.ID,
		Topic:          s.topic,
		Payload: models.JSONMap{
			"id":         notification.ID.String(),
			"user_id":    notification.UserID.String(),
			"type":       notification.Type,
			"channel":    notification.Channel,
			"priority":   notification.Priority,
			"title":      notification.Title,
			"message":    notification.Message,
			"created_at": notification.CreatedAt,
		},
		Published: false,
		CreatedAt: time.Now(),
	}

	if err := s.repository.CreateOutboxEntry(ctx, outboxItem); err != nil {
		return fmt.Errorf("failed to create outbox entry for streak reminder: %w", err)
	}

	return nil
}

// ProcessOutbox processes unpublished outbox items
func (s *notificationService) ProcessOutbox(ctx context.Context) error {
	// Get unpublished outbox items
	outboxItems, err := s.repository.GetUnpublishedOutbox(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get unpublished outbox: %w", err)
	}

	for _, item := range outboxItems {
		// Publish to Kafka
		message := &sarama.ProducerMessage{
			Topic: item.Topic,
			Key:   sarama.StringEncoder(item.NotificationID.String()),
			Value: sarama.ByteEncoder(mustMarshalJSON(item.Payload)),
		}

		partition, offset, err := s.producer.SendMessage(message)
		if err != nil {
			return fmt.Errorf("failed to send message to Kafka: %w", err)
		}

		// Mark as published
		if err := s.repository.MarkOutboxPublished(ctx, item.ID); err != nil {
			return fmt.Errorf("failed to mark outbox as published: %w", err)
		}

		// Log success
		fmt.Printf("Published notification %s to Kafka: partition=%d, offset=%d\n",
			item.NotificationID, partition, offset)
	}

	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
}
