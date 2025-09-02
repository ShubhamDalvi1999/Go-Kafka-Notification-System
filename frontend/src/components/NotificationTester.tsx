import React, { useState, useEffect } from 'react';
import './NotificationTester.css';

interface Notification {
  id: string;
  title: string;
  message: string;
  type: string;
  priority: string;
  isRead: boolean;
  createdAt: string;
  userId: string;
  metadata?: Record<string, any>;
}

export const NotificationTester: React.FC = () => {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Use the same demo user as App.tsx
  const [testUserId] = useState('0241733d-3384-483a-9fd9-5d373e5a2fe6');

  // Fetch notifications from backend
  const fetchNotifications = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(`/api/v1/notifications/${testUserId}`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data = await response.json();
      // Map backend shape { data: Notification[] } to tester Notification[]
      const list = (data.data || data.notifications || []).map((n: any) => ({
        id: n.id,
        title: n.title ?? '',
        message: n.message,
        type: n.type,
        priority: n.priority,
        isRead: Boolean(n.read_at) || n.isRead,
        createdAt: n.created_at || n.createdAt,
        userId: n.user_id || n.userId,
        metadata: n.metadata || {},
      }));
      setNotifications(list);
    } catch (err) {
      setError(`Failed to fetch notifications: ${err instanceof Error ? err.message : 'Unknown error'}`);
      console.error('Error fetching notifications:', err);
    } finally {
      setLoading(false);
    }
  };

  // Generate a test notification
  const generateTestNotification = async (type: string) => {
    try {
      setLoading(true);
      setError(null);
      
      const notificationData = {
        user_id: testUserId,
        type: type,
        channel: 'in_app',
        priority: 'medium',
        title: `Test ${type} Notification`,
        message: `This is a test ${type} notification generated at ${new Date().toLocaleTimeString()}`,
        metadata: {
          test: true,
          generated_at: new Date().toISOString()
        }
      };

      const response = await fetch('/api/v1/notifications', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(notificationData),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      console.log('Notification created:', result);
      
      // Refresh notifications
      await fetchNotifications();
      
    } catch (err) {
      setError(`Failed to create notification: ${err instanceof Error ? err.message : 'Unknown error'}`);
      console.error('Error creating notification:', err);
    } finally {
      setLoading(false);
    }
  };

  // Create daily reminder via generic create endpoint for simplicity
  const createDailyReminder = async () => {
    await generateTestNotification('daily_reminder');
  };

  // Mark notification as read
  const markAsRead = async (notificationId: string) => {
    try {
      const response = await fetch(`/api/v1/notifications/${notificationId}/read`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      // Update local state
      setNotifications(prev => 
        prev.map(notification => 
          notification.id === notificationId 
            ? { ...notification, isRead: true }
            : notification
        )
      );
      
    } catch (err) {
      console.error('Error marking notification as read:', err);
    }
  };

  // Load notifications on component mount
  useEffect(() => {
    fetchNotifications();
  }, []);

  return (
    <div className="notification-tester">
      <h2>🔔 Notification Tester</h2>
      <p>Test user ID: <code>{testUserId}</code></p>
      
      {/* Test Controls */}
      <div className="test-controls">
        <h3>Generate Test Notifications</h3>
        <div className="button-group">
          <button 
            onClick={() => generateTestNotification('daily_reminder')}
            disabled={loading}
            className="test-btn daily"
          >
            📅 Daily Reminder
          </button>
          
          <button 
            onClick={() => generateTestNotification('streak_reminder')}
            disabled={loading}
            className="test-btn streak"
          >
            🔥 Streak Reminder
          </button>
          
          <button 
            onClick={() => generateTestNotification('achievement_unlock')}
            disabled={loading}
            className="test-btn achievement"
          >
            🏆 Achievement
          </button>
          
          <button 
            onClick={() => generateTestNotification('we_miss_you')}
            disabled={loading}
            className="test-btn miss-you"
          >
            💝 We Miss You
          </button>
          
          <button 
            onClick={createDailyReminder}
            disabled={loading}
            className="test-btn scheduler"
          >
            ⏰ Scheduler Reminder
          </button>
        </div>
        
        <button 
          onClick={fetchNotifications}
          disabled={loading}
          className="refresh-btn"
        >
          🔄 Refresh Notifications
        </button>
      </div>

      {/* Error Display */}
      {error && (
        <div className="error-message">
          ❌ {error}
        </div>
      )}

      {/* Notifications List */}
      <div className="notifications-section">
        <h3>📋 Current Notifications ({notifications.length})</h3>
        
        {loading ? (
          <div className="loading">Loading...</div>
        ) : notifications.length === 0 ? (
          <div className="no-notifications">
            No notifications yet. Generate some using the buttons above!
          </div>
        ) : (
          <div className="notifications-list">
            {notifications.map((notification) => (
              <div 
                key={notification.id} 
                className={`notification-item ${notification.isRead ? 'read' : 'unread'}`}
              >
                <div className="notification-header">
                  <span className={`type-badge ${notification.type}`}>
                    {notification.type}
                  </span>
                  <span className={`priority-badge ${notification.priority}`}>
                    {notification.priority}
                  </span>
                  <span className="timestamp">
                    {new Date(notification.createdAt).toLocaleString()}
                  </span>
                </div>
                
                <h4 className="notification-title">{notification.title}</h4>
                <p className="notification-message">{notification.message}</p>
                
                {!notification.isRead && (
                  <button 
                    onClick={() => markAsRead(notification.id)}
                    className="mark-read-btn"
                  >
                    Mark as Read
                  </button>
                )}
                
                {notification.metadata && Object.keys(notification.metadata).length > 0 && (
                  <details className="metadata">
                    <summary>Metadata</summary>
                    <pre>{JSON.stringify(notification.metadata, null, 2)}</pre>
                  </details>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
