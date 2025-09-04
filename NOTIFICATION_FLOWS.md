### Notification Flows (Triggers → Delivery)

This document visualizes how a trigger becomes a user-visible notification without WebSockets. UI refreshes via HTTP reads.

Legend
- Frontend: React app (Vite :3000)
- Producer: API service (:8082)
- DB: Postgres (:5432)
- Kafka: Broker (:9092)
- Consumer: Consumer service (:8081, demo in-memory store)

### 1) Event-Driven: Practice Completed → Achievement
```mermaid
sequenceDiagram
    autonumber
    participant Frontend
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Frontend->>Producer: POST /api/v1/events/practice-completed { user_id, points? }
    Producer->>DB: INSERT notifications (type=achievement_unlock, status=queued)
    Producer->>DB: INSERT outbox_notifications (published=false)
    Note right of Producer: Dev: immediate publish
    Producer->>Kafka: Publish message (topic=notifications)
    Kafka-->>Consumer: Deliver message
    Consumer->>Consumer: Update in-memory per-user list (demo)
    Frontend->>Producer: GET /api/v1/notifications/:userID
    Producer->>DB: SELECT notifications
    DB-->>Producer: Rows
    Producer-->>Frontend: { data: [...] }
```

Implementation snippets

```go
// backend/cmd/producer/main.go
api.POST("/events/practice-completed", handlers.PracticeCompleted)

// backend/pkg/handlers/notification_handlers.go
func (h *NotificationHandlers) PracticeCompleted(c *gin.Context) {
    var req struct{ UserID uuid.UUID `json:"user_id"` }
    _ = c.ShouldBindJSON(&req)
    newReq := &models.CreateNotificationRequest{
        UserID: req.UserID,
        Type:   models.AchievementUnlock,
        Channel: models.ChannelInApp,
    }
    _, _ = h.notificationService.CreateNotification(c.Request.Context(), newReq)
}
```

### 2) Manual Create: API → Any Notification Type
```mermaid
sequenceDiagram
    autonumber
    participant Frontend
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Frontend->>Producer: POST /api/v1/notifications { user_id, type, channel, ... }
    Producer->>DB: INSERT notifications (status=queued)
    Producer->>DB: INSERT outbox_notifications (published=false)
    Producer->>Kafka: Publish (dev immediate) or via outbox processor
    Kafka-->>Consumer: Message
    Consumer->>Consumer: Update in-memory store
    Frontend->>Producer: GET /api/v1/notifications/:userID
    Producer->>DB: SELECT
    DB-->>Producer: Rows
    Producer-->>Frontend: { data: [...] }
```

Implementation snippets

```go
// backend/cmd/producer/main.go
api.POST("/notifications", handlers.CreateNotification)
api.GET("/notifications/:userID", handlers.GetUserNotifications)
api.PUT("/notifications/:id/read", handlers.MarkAsRead)

// backend/pkg/handlers/notification_handlers.go
func (h *NotificationHandlers) CreateNotification(c *gin.Context) {
    var req models.CreateNotificationRequest
    _ = c.ShouldBindJSON(&req)
    n, _ := h.notificationService.CreateNotification(c.Request.Context(), &req)
    c.JSON(http.StatusCreated, gin.H{"data": n})
}
```

### 3) Scheduled: Daily Reminder
```mermaid
sequenceDiagram
    autonumber
    participant Scheduler
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Scheduler->>DB: Query users by preferences (daily, quiet hours, limits)
    Scheduler->>Producer: Create daily_reminder (service call)
    Producer->>DB: INSERT notifications + outbox
    Producer->>Kafka: Publish (dev immediate) or outbox processor
    Kafka-->>Consumer: Message
    Consumer->>Consumer: Update per-user list
```

Implementation snippets

```go
// backend/cmd/scheduler/scheduler.go
go s.startDailyReminderScheduler()

func (s *SchedulerService) processDailyReminders() error {
    ctx := context.Background()
    users, _ := s.getUsersNeedingDailyReminders(ctx)
    for _, u := range users { _ = s.createDailyReminder(ctx, u) }
    return nil
}

func (s *SchedulerService) createDailyReminder(ctx context.Context, user models.User) error {
    n := &models.Notification{Type: models.DailyReminder, Channel: models.ChannelInApp}
    return s.repository.CreateNotification(ctx, n)
}
```

### 4) Scheduled: Streak Reminder (At-Risk)
```mermaid
sequenceDiagram
    autonumber
    participant Scheduler
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Scheduler->>DB: Find users with current_streak>0 AND last_activity_date<today
    Scheduler->>Producer: Create streak_reminder (priority=high)
    Producer->>DB: INSERT notifications + outbox
    Producer->>Kafka: Publish (dev immediate) or outbox processor
    Kafka-->>Consumer: Message
    Consumer->>Consumer: Update per-user list
```

Implementation snippets

```go
// backend/cmd/scheduler/scheduler.go
go s.startStreakReminderScheduler()

func (s *SchedulerService) processStreakReminders() error {
    ctx := context.Background()
    users, _ := s.getUsersNeedingStreakReminders(ctx)
    for _, u := range users { _ = s.createStreakReminder(ctx, u) }
    return nil
}

func (s *SchedulerService) createStreakReminder(ctx context.Context, user models.User) error {
    n := &models.Notification{Type: models.StreakReminder, Channel: models.ChannelInApp}
    return s.repository.CreateNotification(ctx, n)
}
```

### 5) Scheduled: We Miss You (Inactivity)
```mermaid
sequenceDiagram
    autonumber
    participant Scheduler
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Scheduler->>DB: Find users inactive > N days (preferences enabled)
    Scheduler->>Producer: Create we_miss_you (priority=low/medium)
    Producer->>DB: INSERT notifications + outbox
    Producer->>Kafka: Publish (dev immediate) or outbox processor
    Kafka-->>Consumer: Message
    Consumer->>Consumer: Update per-user list
```

Implementation snippets

```go
// backend/cmd/scheduler/scheduler.go
go s.startEngagementNudgeScheduler()

func (s *SchedulerService) processEngagementNudges() error {
    ctx := context.Background()
    users, _ := s.getInactiveUsersForEngagementNudge(ctx)
    for _, u := range users { _ = s.createEngagementNudge(ctx, u) }
    return nil
}

func (s *SchedulerService) createEngagementNudge(ctx context.Context, user models.User) error {
    n := &models.Notification{Type: models.WeMissYou, Channel: models.ChannelInApp}
    return s.repository.CreateNotification(ctx, n)
}
```

### 6) Scheduled: Weekly Recap (Mondays)
```mermaid
sequenceDiagram
    autonumber
    participant Scheduler
    participant Producer
    participant DB
    participant Kafka
    participant Consumer

    Scheduler->>DB: Select active users for weekly recap (Monday)
    Scheduler->>Producer: Create weekly_recap (summary)
    Producer->>DB: INSERT notifications + outbox
    Producer->>Kafka: Publish (dev immediate) or outbox processor
    Kafka-->>Consumer: Message
    Consumer->>Consumer: Update per-user list
```

Implementation snippets

```go
// backend/cmd/scheduler/scheduler.go
go s.startWeeklyRecapScheduler()

func (s *SchedulerService) processWeeklyRecaps() error {
    if time.Now().Weekday() != time.Monday { return nil }
    ctx := context.Background()
    users, _ := s.getActiveUsersForWeeklyRecap(ctx)
    for _, u := range users { _ = s.createWeeklyRecap(ctx, u) }
    return nil
}

func (s *SchedulerService) createWeeklyRecap(ctx context.Context, user models.User) error {
    n := &models.Notification{Type: models.WeeklyRecap, Channel: models.ChannelInApp}
    return s.repository.CreateNotification(ctx, n)
}
```

Notes
- In dev, Producer publishes immediately after outbox insert to minimize latency.
- In prod, prefer background outbox processing, retries, and DLQ.
- Frontend surfaces updates via HTTP polling (`GET /api/v1/notifications/:userID`).


