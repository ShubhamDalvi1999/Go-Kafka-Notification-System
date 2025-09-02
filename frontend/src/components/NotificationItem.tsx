import React, { useState } from 'react';
import { Notification as NotificationType } from '../types/notification';
import './NotificationItem.css';

interface NotificationItemProps {
  notification: NotificationType;
  onMarkAsRead: () => void;
}

export const NotificationItem: React.FC<NotificationItemProps> = ({
  notification,
  onMarkAsRead
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleToggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  const handleMarkAsRead = (e: React.MouseEvent) => {
    e.stopPropagation();
    onMarkAsRead();
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'low': return 'priority-low';
      case 'medium': return 'priority-medium';
      case 'high': return 'priority-high';
      case 'urgent': return 'priority-urgent';
      default: return 'priority-medium';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'daily_reminder': return '📅';
      case 'streak_reminder': return '🔥';
      case 'last_chance_alert': return '⚠️';
      case 'achievement_unlock': return '🏆';
      case 'xp_goal_reminder': return '⭐';
      case 'league_update': return '🏁';
      case 'we_miss_you': return '💔';
      case 'event_notification': return '🎉';
      case 'new_course': return '📚';
      case 'practice_needed': return '💪';
      case 'weekly_recap': return '📊';
      default: return '🔔';
    }
  };

  const formatTime = (date: Date) => {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    
    return date.toLocaleDateString();
  };

  const getChannelIcon = (channel: string) => {
    switch (channel) {
      case 'in_app': return '📱';
      case 'push': return '📲';
      case 'email': return '📧';
      case 'sms': return '💬';
      default: return '📱';
    }
  };

  return (
    <div 
      className={`notification-item ${notification.isRead ? 'read' : 'unread'} ${getPriorityColor(notification.priority)}`}
      onClick={handleToggleExpand}
    >
      {/* Header */}
      <div className="notification-header">
        <div className="notification-icon">
          {getTypeIcon(notification.type)}
        </div>
        
        <div className="notification-content">
          <div className="notification-title">
            {notification.title || notification.type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
          </div>
          
          <div className="notification-meta">
            <span className="notification-time">
              {formatTime(notification.createdAt)}
            </span>
            <span className="notification-channel">
              {getChannelIcon(notification.channel)}
            </span>
            {!notification.isRead && (
              <span className="unread-indicator">•</span>
            )}
          </div>
        </div>
        
        <div className="notification-actions">
          {!notification.isRead && (
            <button
              className="mark-read-button"
              onClick={handleMarkAsRead}
              title="Mark as read"
            >
              ✓
            </button>
          )}
          
          <button
            className={`expand-button ${isExpanded ? 'expanded' : ''}`}
            onClick={handleToggleExpand}
            title={isExpanded ? 'Collapse' : 'Expand'}
          >
            {isExpanded ? '−' : '+'}
          </button>
        </div>
      </div>

      {/* Message */}
      <div className="notification-message">
        {notification.message}
      </div>

      {/* Expanded Content */}
      {isExpanded && (
        <div className="notification-expanded">
          <div className="notification-details">
            <div className="detail-row">
              <span className="detail-label">Type:</span>
              <span className="detail-value">
                {notification.type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
              </span>
            </div>
            
            <div className="detail-row">
              <span className="detail-label">Priority:</span>
              <span className={`detail-value priority-${notification.priority}`}>
                {notification.priority.charAt(0).toUpperCase() + notification.priority.slice(1)}
              </span>
            </div>
            
            <div className="detail-row">
              <span className="detail-label">Channel:</span>
              <span className="detail-value">
                {notification.channel.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
              </span>
            </div>
            
            {notification.scheduledFor && (
              <div className="detail-row">
                <span className="detail-label">Scheduled for:</span>
                <span className="detail-value">
                  {notification.scheduledFor.toLocaleString()}
                </span>
              </div>
            )}
            
            {notification.metadata && Object.keys(notification.metadata).length > 0 && (
              <div className="detail-row">
                <span className="detail-label">Metadata:</span>
                <span className="detail-value">
                  <pre className="metadata-json">
                    {JSON.stringify(notification.metadata, null, 2)}
                  </pre>
                </span>
              </div>
            )}
          </div>
          
          {/* Action Buttons */}
          <div className="notification-actions-expanded">
            {notification.type === 'daily_reminder' && (
              <button className="action-button primary">
                Start Practice
              </button>
            )}
            
            {notification.type === 'streak_reminder' && (
              <button className="action-button primary">
                Continue Streak
              </button>
            )}
            
            {notification.type === 'achievement_unlock' && (
              <button className="action-button secondary">
                View Achievement
              </button>
            )}
            
            <button className="action-button secondary">
              Dismiss
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
