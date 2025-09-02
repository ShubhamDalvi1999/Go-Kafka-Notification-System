package services

import (
	"context"
	"testing"
	"time"

	"kafka-notify/pkg/models"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetNotificationByID(ctx context.Context, notificationID uuid.UUID) (*models.Notification, error) {
	args := m.Called(ctx, notificationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsDelivered(ctx context.Context, notificationID uuid.UUID) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsSent(ctx context.Context, notificationID uuid.UUID) error {
	args := m.Called(ctx, notificationID)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUnpublishedOutbox(ctx context.Context, limit int) ([]models.OutboxNotification, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]models.OutboxNotification), args.Error(1)
}

func (m *MockNotificationRepository) MarkOutboxPublished(ctx context.Context, outboxID int64) error {
	args := m.Called(ctx, outboxID)
	return args.Error(0)
}

func (m *MockNotificationRepository) CreateOutboxEntry(ctx context.Context, outboxItem *models.OutboxNotification) error {
	args := m.Called(ctx, outboxItem)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.UserNotificationPreferences), args.Error(1)
}

func (m *MockNotificationRepository) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, prefs *models.UserNotificationPreferences) error {
	args := m.Called(ctx, userID, prefs)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUserEngagementStreak(ctx context.Context, userID uuid.UUID, streakType string) (*models.UserEngagementStreak, error) {
	args := m.Called(ctx, userID, streakType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEngagementStreak), args.Error(1)
}

func (m *MockNotificationRepository) UpdateUserEngagementStreak(ctx context.Context, streak *models.UserEngagementStreak) error {
	args := m.Called(ctx, streak)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationsByStatus(ctx context.Context, status models.DeliveryStatus, limit int) ([]models.Notification, error) {
	args := m.Called(ctx, status, limit)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetScheduledNotifications(ctx context.Context, before time.Time, limit int) ([]models.Notification, error) {
	args := m.Called(ctx, before, limit)
	return args.Get(0).([]models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) CreateDeliveryAttempt(ctx context.Context, attempt *models.NotificationDeliveryAttempt) error {
	args := m.Called(ctx, attempt)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetNotificationTemplates(ctx context.Context, notificationType models.NotificationType, channel models.NotificationChannel) ([]models.NotificationTemplate, error) {
	args := m.Called(ctx, notificationType, channel)
	return args.Get(0).([]models.NotificationTemplate), args.Error(1)
}

// MockKafkaProducer is a mock implementation of sarama.SyncProducer
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	args := m.Called(msg)
	return int32(args.Int(0)), args.Int64(1), args.Error(2)
}

func (m *MockKafkaProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreateNotification_ValidRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotificationRepository)
	mockProducer := new(MockKafkaProducer)

	service := NewNotificationService(mockRepo, mockProducer, "test-topic")

	req := &models.CreateNotificationRequest{
		UserID:   uuid.New(),
		Type:     models.DailyReminder,
		Channel:  models.ChannelInApp,
		Priority: models.PriorityMedium,
		Message:  "Test notification",
	}

	ctx := context.Background()

	// Mock expectations
	mockRepo.On("CreateNotification", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)
	mockRepo.On("CreateOutboxEntry", ctx, mock.AnythingOfType("*models.OutboxNotification")).Return(nil)

	// Act
	notification, err := service.CreateNotification(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, req.UserID, notification.UserID)
	assert.Equal(t, req.Type, notification.Type)
	assert.Equal(t, req.Channel, notification.Channel)
	assert.Equal(t, req.Priority, notification.Priority)
	assert.Equal(t, req.Message, notification.Message)
	assert.Equal(t, models.StatusQueued, notification.Status)

	mockRepo.AssertExpectations(t)
}

func TestCreateNotification_InvalidType(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotificationRepository)
	mockProducer := new(MockKafkaProducer)

	service := NewNotificationService(mockRepo, mockProducer, "test-topic")

	req := &models.CreateNotificationRequest{
		UserID:  uuid.New(),
		Type:    "invalid_type",
		Channel: models.ChannelInApp,
		Message: "Test notification",
	}

	ctx := context.Background()

	// Act
	notification, err := service.CreateNotification(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, notification)
	assert.Contains(t, err.Error(), "invalid notification type")
}

func TestCreateNotification_InvalidChannel(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotificationRepository)
	mockProducer := new(MockKafkaProducer)

	service := NewNotificationService(mockRepo, mockProducer, "test-topic")

	req := &models.CreateNotificationRequest{
		UserID:  uuid.New(),
		Type:    models.DailyReminder,
		Channel: "invalid_channel",
		Message: "Test notification",
	}

	ctx := context.Background()

	// Act
	notification, err := service.CreateNotification(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, notification)
	assert.Contains(t, err.Error(), "invalid notification channel")
}

func TestGetUserNotifications_ValidRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotificationRepository)
	mockProducer := new(MockKafkaProducer)

	service := NewNotificationService(mockRepo, mockProducer, "test-topic")

	userID := uuid.New()
	ctx := context.Background()
	limit := 10
	offset := 0

	expectedNotifications := []models.Notification{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      models.DailyReminder,
			Channel:   models.ChannelInApp,
			Message:   "Test notification 1",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      models.StreakReminder,
			Channel:   models.ChannelInApp,
			Message:   "Test notification 2",
			CreatedAt: time.Now(),
		},
	}

	// Mock expectations
	mockRepo.On("GetUserNotifications", ctx, userID, limit, offset).Return(expectedNotifications, nil)

	// Act
	notifications, err := service.GetUserNotifications(ctx, userID, limit, offset)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, notifications, 2)
	assert.Equal(t, expectedNotifications, notifications)

	mockRepo.AssertExpectations(t)
}

func TestMarkAsRead_ValidRequest(t *testing.T) {
	// Arrange
	mockRepo := new(MockNotificationRepository)
	mockProducer := new(MockKafkaProducer)

	service := NewNotificationService(mockRepo, mockProducer, "test-topic")

	notificationID := uuid.New()
	ctx := context.Background()

	// Mock expectations
	mockRepo.On("MarkAsRead", ctx, notificationID).Return(nil)

	// Act
	err := service.MarkAsRead(ctx, notificationID)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
