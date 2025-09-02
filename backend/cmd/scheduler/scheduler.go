package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kafka-notify/pkg/models"
	"kafka-notify/pkg/repository"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	DBConnectionString = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	CheckInterval      = 5 * time.Minute // Check every 5 minutes instead of every minute
)

// SchedulerService handles automated notification scheduling
type SchedulerService struct {
	repository repository.NotificationRepository
	stopChan   chan os.Signal
	db         *sql.DB
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService() (*SchedulerService, error) {
	// Initialize database connection
	db, err := sql.Open("postgres", DBConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Cap pool size for pooler (Supabase/pgbouncer)
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize repository
	repo := repository.NewPostgresNotificationRepository(db)

	service := &SchedulerService{
		repository: repo,
		stopChan:   make(chan os.Signal, 1),
		db:         db,
	}

	return service, nil
}

// Start starts the scheduler service
func (s *SchedulerService) Start() error {
	log.Println("Starting notification scheduler service...")

	// Start background schedulers
	go s.startDailyReminderScheduler()
	go s.startStreakReminderScheduler()
	go s.startWeeklyRecapScheduler()
	go s.startEngagementNudgeScheduler()

	log.Println("Scheduler service started successfully")

	// Wait for shutdown signal
	signal.Notify(s.stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-s.stopChan

	log.Println("Shutting down scheduler service...")
	return s.Shutdown()
}

// startDailyReminderScheduler starts the daily reminder scheduler
func (s *SchedulerService) startDailyReminderScheduler() {
	ticker := time.NewTicker(CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processDailyReminders(); err != nil {
				log.Printf("Daily reminder scheduler error: %v", err)
			}
		case <-s.stopChan:
			return
		}
	}
}

// startStreakReminderScheduler starts the streak reminder scheduler
func (s *SchedulerService) startStreakReminderScheduler() {
	ticker := time.NewTicker(CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processStreakReminders(); err != nil {
				log.Printf("Streak reminder scheduler error: %v", err)
			}
		case <-s.stopChan:
			return
		}
	}
}

// startWeeklyRecapScheduler starts the weekly recap scheduler
func (s *SchedulerService) startWeeklyRecapScheduler() {
	ticker := time.NewTicker(24 * time.Hour) // Check once per day
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processWeeklyRecaps(); err != nil {
				log.Printf("Weekly recap scheduler error: %v", err)
			}
		case <-s.stopChan:
			return
		}
	}
}

// startEngagementNudgeScheduler starts the engagement nudge scheduler
func (s *SchedulerService) startEngagementNudgeScheduler() {
	ticker := time.NewTicker(6 * time.Hour) // Check every 6 hours
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processEngagementNudges(); err != nil {
				log.Printf("Engagement nudge scheduler error: %v", err)
			}
		case <-s.stopChan:
			return
		}
	}
}

// processDailyReminders processes daily reminders for all users
func (s *SchedulerService) processDailyReminders() error {
	ctx := context.Background()

	// Get all users who need daily reminders
	users, err := s.getUsersNeedingDailyReminders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users needing daily reminders: %w", err)
	}

	if len(users) > 0 {
		log.Printf("Processing daily reminders for %d users", len(users))
	}

	for _, user := range users {
		if err := s.createDailyReminder(ctx, user); err != nil {
			log.Printf("Failed to create daily reminder for user %s: %v", user.ID, err)
			continue
		}
	}

	return nil
}

// processStreakReminders processes streak reminders for users at risk
func (s *SchedulerService) processStreakReminders() error {
	ctx := context.Background()

	// Get users who need streak reminders
	users, err := s.getUsersNeedingStreakReminders(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users needing streak reminders: %w", err)
	}

	if len(users) > 0 {
		log.Printf("Processing streak reminders for %d users", len(users))
	}

	for _, user := range users {
		if err := s.createStreakReminder(ctx, user); err != nil {
			log.Printf("Failed to create streak reminder for user %s: %v", user.ID, err)
			continue
		}
	}

	return nil
}

// processWeeklyRecaps processes weekly recaps for active users
func (s *SchedulerService) processWeeklyRecaps() error {
	ctx := context.Background()
	now := time.Now()

	// Only send weekly recaps on Mondays
	if now.Weekday() != time.Monday {
		return nil
	}

	// Get active users for weekly recap
	users, err := s.getActiveUsersForWeeklyRecap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active users for weekly recap: %w", err)
	}

	if len(users) > 0 {
		log.Printf("Processing weekly recaps for %d users", len(users))
	}

	for _, user := range users {
		if err := s.createWeeklyRecap(ctx, user); err != nil {
			log.Printf("Failed to create weekly recap for user %s: %v", user.ID, err)
			continue
		}
	}

	return nil
}

// processEngagementNudges processes engagement nudges for inactive users
func (s *SchedulerService) processEngagementNudges() error {
	ctx := context.Background()

	// Get inactive users who need engagement nudges
	users, err := s.getInactiveUsersForEngagementNudge(ctx)
	if err != nil {
		return fmt.Errorf("failed to get inactive users for engagement nudge: %w", err)
	}

	if len(users) > 0 {
		log.Printf("Processing engagement nudges for %d users", len(users))
	}

	for _, user := range users {
		if err := s.createEngagementNudge(ctx, user); err != nil {
			log.Printf("Failed to create engagement nudge for user %s: %v", user.ID, err)
			continue
		}
	}

	return nil
}

// getUsersNeedingDailyReminders gets users who need daily reminders
func (s *SchedulerService) getUsersNeedingDailyReminders(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT DISTINCT u.user_id, u.name, u.email
		FROM users u
		JOIN user_notification_preferences unp ON u.user_id = unp.user_id
		WHERE unp.type = 'daily_reminder' 
		  AND unp.channel = 'in_app' 
		  AND unp.enabled = true
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n 
			WHERE n.user_id = u.user_id 
			  AND n.type = 'daily_reminder' 
			  AND n.created_at::date = current_date
		  )
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users needing daily reminders: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// getUsersNeedingStreakReminders gets users who need streak reminders
func (s *SchedulerService) getUsersNeedingStreakReminders(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT DISTINCT u.user_id, u.name, u.email
		FROM users u
		JOIN user_notification_preferences unp ON u.user_id = unp.user_id
		JOIN user_engagement_streaks ues ON u.user_id = ues.user_id
		WHERE unp.type = 'streak_reminder' 
		  AND unp.channel = 'in_app' 
		  AND unp.enabled = true
		  AND ues.streak_type = 'practice'
		  AND ues.current_streak > 0
		  AND ues.last_activity_date < current_date
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n 
			WHERE n.user_id = u.user_id 
			  AND n.type = 'streak_reminder' 
			  AND n.created_at::date = current_date
		  )
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users needing streak reminders: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// getActiveUsersForWeeklyRecap gets active users for weekly recap
func (s *SchedulerService) getActiveUsersForWeeklyRecap(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT DISTINCT u.user_id, u.name, u.email
		FROM users u
		JOIN user_notification_preferences unp ON u.user_id = unp.user_id
		WHERE unp.type = 'weekly_recap' 
		  AND unp.channel = 'in_app' 
		  AND unp.enabled = true
		  AND EXISTS (
			SELECT 1 FROM user_engagement_streaks ues 
			WHERE ues.user_id = u.user_id 
			  AND ues.streak_type = 'practice'
			  AND ues.current_streak > 0
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n 
			WHERE n.user_id = u.user_id 
			  AND n.type = 'weekly_recap' 
			  AND n.created_at >= date_trunc('week', current_date)
		  )
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users for weekly recap: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// getInactiveUsersForEngagementNudge gets inactive users for engagement nudge
func (s *SchedulerService) getInactiveUsersForEngagementNudge(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT DISTINCT u.user_id, u.name, u.email
		FROM users u
		JOIN user_notification_preferences unp ON u.user_id = unp.user_id
		WHERE unp.type = 'we_miss_you' 
		  AND unp.channel = 'in_app' 
		  AND unp.enabled = true
		  AND EXISTS (
			SELECT 1 FROM user_engagement_streaks ues 
			WHERE ues.user_id = u.user_id 
			  AND ues.streak_type = 'practice'
			  AND ues.last_activity_date < current_date - interval '7 days'
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM notifications n 
			WHERE n.user_id = u.user_id 
			  AND n.type = 'we_miss_you' 
			  AND n.created_at >= current_date - interval '7 days'
		  )
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query inactive users for engagement nudge: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Failed to scan user: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// createDailyReminder creates a daily reminder for a user
func (s *SchedulerService) createDailyReminder(ctx context.Context, user models.User) error {
	// Get user engagement streak
	streak, err := s.repository.GetUserEngagementStreak(ctx, user.ID, "practice")
	if err != nil {
		log.Printf("Failed to get user streak for %s: %v", user.ID, err)
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
		Topic:          "notifications",
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
		log.Printf("Failed to create outbox entry for daily reminder: %v", err)
	}

	log.Printf("Created daily reminder for user %s (streak: %d)", user.ID, currentStreak)
	return nil
}

// createStreakReminder creates a streak reminder for a user
func (s *SchedulerService) createStreakReminder(ctx context.Context, user models.User) error {
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
		Topic:          "notifications",
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
		log.Printf("Failed to create outbox entry for streak reminder: %v", err)
	}

	log.Printf("Created streak reminder for user %s (streak: %d)", user.ID, streak.CurrentStreak)
	return nil
}

// createWeeklyRecap creates a weekly recap for a user
func (s *SchedulerService) createWeeklyRecap(ctx context.Context, user models.User) error {
	// Get user engagement streak
	streak, err := s.repository.GetUserEngagementStreak(ctx, user.ID, "practice")
	if err != nil {
		log.Printf("Failed to get user streak for weekly recap: %v", err)
		// Continue with default values
	}

	currentStreak := 0
	if streak != nil {
		currentStreak = streak.CurrentStreak
	}

	// Create weekly recap notification
	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      models.WeeklyRecap,
		Channel:   models.ChannelInApp,
		Priority:  models.PriorityLow,
		Title:     stringPtr("Your Weekly Progress Report"),
		Message:   fmt.Sprintf("Great week %s! You maintained your %d-day streak! Keep up the amazing work! 🎉", user.Name, currentStreak),
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repository.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create weekly recap: %w", err)
	}

	// Create outbox entry
	outboxItem := &models.OutboxNotification{
		NotificationID: notification.ID,
		Topic:          "notifications",
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
		log.Printf("Failed to create outbox entry for weekly recap: %v", err)
	}

	log.Printf("Created weekly recap for user %s", user.ID)
	return nil
}

// createEngagementNudge creates an engagement nudge for a user
func (s *SchedulerService) createEngagementNudge(ctx context.Context, user models.User) error {
	// Create engagement nudge notification
	notification := &models.Notification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      models.WeMissYou,
		Channel:   models.ChannelInApp,
		Priority:  models.PriorityLow,
		Title:     stringPtr("We Miss You!"),
		Message:   fmt.Sprintf("Hey %s! It's been a while since your last practice. Your skills are getting rusty! Come back and practice! 💪", user.Name),
		Status:    models.StatusQueued,
		CreatedAt: time.Now(),
	}

	// Save to database
	if err := s.repository.CreateNotification(ctx, notification); err != nil {
		return fmt.Errorf("failed to create engagement nudge: %w", err)
	}

	// Create outbox entry
	outboxItem := &models.OutboxNotification{
		NotificationID: notification.ID,
		Topic:          "notifications",
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
		log.Printf("Failed to create outbox entry for engagement nudge: %v", err)
	}

	log.Printf("Created engagement nudge for user %s", user.ID)
	return nil
}

// Shutdown gracefully shuts down the service
func (s *SchedulerService) Shutdown() error {
	log.Println("Shutting down scheduler service...")

	// Close database connection
	if err := s.db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Scheduler service shutdown complete")
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func main() {
	service, err := NewSchedulerService()
	if err != nil {
		log.Fatalf("Failed to create scheduler service: %v", err)
	}

	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start scheduler service: %v", err)
	}
}
