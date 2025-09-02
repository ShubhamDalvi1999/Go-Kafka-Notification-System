package handlers

import (
	"net/http"
	"strconv"

	"kafka-notify/internal/services"
	"kafka-notify/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandlers handles HTTP requests for notifications
type NotificationHandlers struct {
	notificationService services.NotificationService
}

// NewNotificationHandlers creates new notification handlers
func NewNotificationHandlers(notificationService services.NotificationService) *NotificationHandlers {
	return &NotificationHandlers{
		notificationService: notificationService,
	}
}

// CreateNotification handles POST /notifications
func (h *NotificationHandlers) CreateNotification(c *gin.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	notification, err := h.notificationService.CreateNotification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification created successfully",
		"data":    notification,
	})
}

// PracticeCompleted handles POST /events/practice-completed
// Simplified event-to-notification mapping for POC
func (h *NotificationHandlers) PracticeCompleted(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Points *int      `json:"points"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	title := ptr("Practice Completed!")
	message := "Great job on completing your practice session. Keep it up!"
	if req.Points != nil {
		message = message + " You earned " + strconv.Itoa(*req.Points) + " XP."
	}

	newReq := &models.CreateNotificationRequest{
		UserID:   req.UserID,
		Type:     models.AchievementUnlock,
		Channel:  models.ChannelInApp,
		Priority: models.PriorityMedium,
		Title:    title,
		Message:  message,
		Metadata: models.JSONMap{"event": "practice_completed"},
	}

	n, err := h.notificationService.CreateNotification(c.Request.Context(), newReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create event notification",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Event notification created",
		"data":    n,
	})
}

func ptr(s string) *string { return &s }

// GetUserNotifications handles GET /notifications/:userID
func (h *NotificationHandlers) GetUserNotifications(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid offset parameter",
		})
		return
	}

	notifications, err := h.notificationService.GetUserNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve notifications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": notifications,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(notifications),
		},
	})
}

// MarkAsRead handles PUT /notifications/:id/read
func (h *NotificationHandlers) MarkAsRead(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid notification ID format",
		})
		return
	}

	if err := h.notificationService.MarkAsRead(c.Request.Context(), notificationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read successfully",
	})
}

// UpdateUserPreferences handles PUT /preferences/:userID
func (h *NotificationHandlers) UpdateUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var prefs models.UserNotificationPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.notificationService.UpdateUserPreferences(c.Request.Context(), userID, &prefs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user preferences",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User preferences updated successfully",
	})
}

// GetUserPreferences handles GET /preferences/:userID
func (h *NotificationHandlers) GetUserPreferences(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	preferences, err := h.notificationService.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve user preferences",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": preferences,
	})
}

// CreateDailyReminder handles POST /reminders/daily
func (h *NotificationHandlers) CreateDailyReminder(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.notificationService.CreateDailyReminder(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create daily reminder",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Daily reminder created successfully",
	})
}

// CreateStreakReminder handles POST /reminders/streak
func (h *NotificationHandlers) CreateStreakReminder(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if err := h.notificationService.CreateStreakReminder(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create streak reminder",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Streak reminder created successfully",
	})
}

// ProcessOutbox handles POST /outbox/process
func (h *NotificationHandlers) ProcessOutbox(c *gin.Context) {
	if err := h.notificationService.ProcessOutbox(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process outbox",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Outbox processed successfully",
	})
}
